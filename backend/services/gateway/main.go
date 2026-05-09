package main

import (
	"MathOverflow/common/client"
	"MathOverflow/common/config"
	"MathOverflow/common/middleware"
	"MathOverflow/common/utils"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func newReverseProxy(target string) *httputil.ReverseProxy {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("invalid target url %s: %v", target, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	originalDirector := proxy.Director

	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// 这里可以统一加一些网关头
		// req.Header.Set("X-Gateway", "forum-gateway")

		// 保留原始 Host 信息也可以根据需要改
		req.Host = targetURL.Host
	}

	proxy.Transport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ResponseHeaderTimeout: 5 * time.Second,
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error":"bad gateway"}`))
	}

	return proxy
}

func main() {
	utils.InitLogger("gateway")

	dir, _ := os.Getwd()
	utils.Logger().WithField("working_dir", dir).Info("gateway starting")
	// 加载配置

	cfg, err := config.LoadConfig("config/gateway.yaml")
	if err != nil {
		utils.Logger().WithError(err).Fatal("failed to load gateway.yaml")
	}

	// 初始化redis数据库
	rdb, err := client.InitRedis(cfg.Redis)
	if err != nil {
		panic(err)
	}

	// 创建两个下游服务代理
	userProxy := newReverseProxy("http://localhost:8081")
	forumProxy := newReverseProxy("http://localhost:8082")

	cors := middleware.CorsMiddleware([]string{"http://localhost:5173", "https://math-overflow.edu"})

	r := gin.Default()
	r.Use(cors)
	r = SetupUserRouter(rdb, r, userProxy)
	r = SetupForumRouter(rdb, r, forumProxy)

	port := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("gateway listening on %s", port)
	r.Run(port)
}
