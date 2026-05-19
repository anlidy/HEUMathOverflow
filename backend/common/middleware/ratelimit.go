package middleware

import (
	"MathOverflow/common/utils"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func normalizeRateLimitPathScope(pathScope string) string {
	if pathScope == "" {
		return "/"
	}
	if pathScope == "/" {
		return pathScope
	}
	return strings.TrimRight(pathScope, "/")
}

func matchRateLimitPathScope(requestPath, pathScope string) bool {
	if pathScope == "/" {
		return true
	}
	if requestPath == pathScope {
		return true
	}
	return strings.HasPrefix(requestPath, pathScope+"/")
}

const redisTokenBucketScript = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local capacity = tonumber(ARGV[3])
local ttl = tonumber(ARGV[4])

local data = redis.call("HMGET", key, "tokens", "ts")
local tokens = tonumber(data[1])
local ts = tonumber(data[2])

if not tokens then
    tokens = capacity
end

if not ts then
    ts = now
end

local timedelta = math.max(0, now - ts)
local refill = timedelta * rate / 1000
tokens = math.min(capacity, tokens + refill)

local allowed = 0
if tokens >= 1 then
    tokens = tokens - 1
    allowed = 1
end

redis.call("HSET", key, "tokens", tokens, "ts", now)
redis.call("PEXPIRE", key, ttl)

return {allowed, tokens}
`

// RedisIPRateLimitMiddleware enforces a distributed token bucket limit per client IP
// for requests under the given path scope.
func RedisIPRateLimitMiddleware(rdb *redis.Client, pathScope string, qps float64, capacity int) gin.HandlerFunc {
	if qps <= 0 {
		qps = 100
	}
	if capacity <= 0 {
		capacity = int(qps)
	}
	pathScope = normalizeRateLimitPathScope(pathScope)

	// 令牌桶过期时间
	ttl := time.Duration(float64(capacity) / qps * 2 * float64(time.Second)) // 2倍时间
	if ttl < time.Minute {
		ttl = time.Minute
	}

	return func(c *gin.Context) {
		if !matchRateLimitPathScope(c.Request.URL.Path, pathScope) {
			c.Next()
			return
		}

		ctx := c.Request.Context()
		key := fmt.Sprintf("ratelimit:path:%s:ip:%s", pathScope, c.ClientIP())
		result, err := rdb.Eval(
			ctx,
			redisTokenBucketScript,
			[]string{key},
			time.Now().UnixMilli(),
			strconv.FormatFloat(qps, 'f', -1, 64),
			capacity,
			ttl.Milliseconds(),
		).Result()
		if err != nil {
			utils.WithContext(ctx).WithError(err).WithFields(map[string]any{
				"client_ip":  c.ClientIP(),
				"limit_qps":  qps,
				"capacity":   capacity,
				"path_scope": pathScope,
			}).Error("redis rate limiter eval failed")
			c.Next()
			return
		}

		values, ok := result.([]any)
		if !ok || len(values) == 0 {
			utils.WithContext(ctx).WithField("result_type", fmt.Sprintf("%T", result)).WithFields(map[string]any{
				"client_ip":  c.ClientIP(),
				"limit_qps":  qps,
				"capacity":   capacity,
				"path_scope": pathScope,
			}).Error("redis rate limiter returned unexpected result")
			c.Next()
			return
		}

		allowed, ok := values[0].(int64)
		if !ok {
			utils.WithContext(ctx).WithField("allowed_type", fmt.Sprintf("%T", values[0])).WithFields(map[string]any{
				"client_ip":  c.ClientIP(),
				"limit_qps":  qps,
				"capacity":   capacity,
				"path_scope": pathScope,
			}).Error("redis rate limiter returned invalid allowed flag")
			c.Next()
			return
		}

		if allowed != 1 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"message": "请求过于频繁，请稍后再试",
				"code":    http.StatusTooManyRequests,
			})
			return
		}
		c.Next()
	}
}
