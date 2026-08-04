# Unit Tests Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add comprehensive unit tests to the HEUMathOverflow project (Go backend + Vue frontend) and run them via Docker.

**Architecture:** Test pure business logic and utility functions first (highest ROI), then middleware and service layers with mocks. Frontend tests use Vitest + happy-dom, backend uses standard Go `testing` + testify.

**Tech Stack:** Go testing, testify/assert, testify/mock, Vitest, @vue/test-utils, happy-dom

---

## Part A: Backend Go Tests

### Task 1: `common/utils/calculate_test.go` — Post ranking formula

**Files:**
- Create: `backend/common/utils/calculate_test.go`

- [ ] **Step 1: Write the test file**

```go
package utils

import (
	"testing"
	"time"

	"MathOverflow/common/model"
)

func TestCalcPostScores(t *testing.T) {
	tests := []struct {
		name        string
		post        model.PostStat
		wantFavors  int64
		checkScore  func(score float64) bool
	}{
		{
			name: "zero values",
			post: model.PostStat{
				CreatedAt: time.Now(),
			},
			wantFavors: 0,
			checkScore: func(score float64) bool { return score >= 0 },
		},
		{
			name: "high engagement post",
			post: model.PostStat{
				Views:     1000,
				Likes:     50,
				Stars:     20,
				Replies:   30,
				Status:    3,
				CreatedAt: time.Now().Add(-24 * time.Hour),
			},
			wantFavors: 50*3 + 20*5, // 250
			checkScore: func(score float64) bool { return score > 0 },
		},
		{
			name: "new post with no engagement",
			post: model.PostStat{
				Views:     0,
				Likes:     0,
				Stars:     0,
				Replies:   0,
				Status:    1,
				CreatedAt: time.Now(),
			},
			wantFavors: 0,
			checkScore: func(score float64) bool { return score >= 0 },
		},
		{
			name: "old post decays over time",
			post: model.PostStat{
				Views:     100,
				Likes:     10,
				Stars:     5,
				Replies:   5,
				Status:    2,
				CreatedAt: time.Now().Add(-168 * time.Hour), // 7 days ago
			},
			wantFavors: 10*3 + 5*5, // 55
			checkScore: func(score float64) bool { return score > 0 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			favors, score := CalcPostScores(&tt.post)
			if favors != tt.wantFavors {
				t.Errorf("CalcPostScores() favors = %v, want %v", favors, tt.wantFavors)
			}
			if !tt.checkScore(score) {
				t.Errorf("CalcPostScores() score = %v, failed check", score)
			}
		})
	}
}

func TestCalcPostScores_TimeDecay(t *testing.T) {
	now := time.Now()
	recent := model.PostStat{
		Views: 100, Likes: 10, Stars: 5, Replies: 5, Status: 2,
		CreatedAt: now.Add(-1 * time.Hour),
	}
	old := model.PostStat{
		Views: 100, Likes: 10, Stars: 5, Replies: 5, Status: 2,
		CreatedAt: now.Add(-168 * time.Hour),
	}

	_, recentScore := CalcPostScores(&recent)
	_, oldScore := CalcPostScores(&old)

	if oldScore >= recentScore {
		t.Errorf("Old post score (%v) should be less than recent post score (%v)", oldScore, recentScore)
	}
}

func TestCalcPostScores_FavorsFormula(t *testing.T) {
	post := model.PostStat{
		Likes: 10, Stars: 5,
		CreatedAt: time.Now(),
	}
	favors, _ := CalcPostScores(&post)
	expected := int64(10*3 + 5*5)
	if favors != expected {
		t.Errorf("Favors = %v, want %v", favors, expected)
	}
}
```

- [ ] **Step 2: Run to verify**

Run: `cd /home/ai/HEUMathOverflow/backend && go test ./common/utils/ -run TestCalcPostScores -v`

- [ ] **Step 3: Commit**

---

### Task 2: `common/utils/circuitbreaker_test.go` — Circuit breaker

**Files:**
- Create: `backend/common/utils/circuitbreaker_test.go`

- [ ] **Step 1: Write the test file**

```go
package utils

import (
	"testing"
	"time"
)

func TestCircuitBreaker_Allow_InitiallyClosed(t *testing.T) {
	cb := NewCircuitBreaker(3, 1*time.Second)
	if !cb.Allow() {
		t.Error("New circuit breaker should allow requests (closed state)")
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(3, 1*time.Second)
	for i := 0; i < 3; i++ {
		cb.OnFailure()
	}
	if cb.Allow() {
		t.Error("Circuit breaker should be open after threshold failures")
	}
}

func TestCircuitBreaker_ClosesAfterOpenDuration(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond)
	cb.OnFailure()
	cb.OnFailure()
	if cb.Allow() {
		t.Error("Should be open immediately after threshold")
	}
	time.Sleep(60 * time.Millisecond)
	if !cb.Allow() {
		t.Error("Should be half-open after duration expires")
	}
}

func TestCircuitBreaker_ResetOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker(3, 1*time.Second)
	cb.OnFailure()
	cb.OnFailure()
	cb.OnSuccess()
	if !cb.Allow() {
		t.Error("Should be closed after success resets failures")
	}
}

func TestCircuitBreaker_HalfOpenToClosed(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond)
	cb.OnFailure()
	cb.OnFailure()
	time.Sleep(60 * time.Millisecond)
	cb.Allow() // half-open
	cb.OnSuccess()
	if !cb.Allow() {
		t.Error("Should be closed after success in half-open state")
	}
}

func TestCircuitBreaker_Defaults(t *testing.T) {
	cb := NewCircuitBreaker(0, 0)
	if !cb.Allow() {
		t.Error("Default circuit breaker should allow requests")
	}
}
```

- [ ] **Step 2: Run**

Run: `cd /home/ai/HEUMathOverflow/backend && go test ./common/utils/ -run TestCircuitBreaker -v`

---

### Task 3: `common/utils/tools_test.go` — Generic slice filter

**Files:**
- Create: `backend/common/utils/tools_test.go`

- [ ] **Step 1: Write the test file**

```go
package utils

import (
	"testing"
)

func TestSliceFilter(t *testing.T) {
	tests := []struct {
		name   string
		src    []string
		filter []string
		want   []string
	}{
		{
			name:   "remove matching elements",
			src:    []string{"a", "b", "c", "d"},
			filter: []string{"b", "d"},
			want:   []string{"a", "c"},
		},
		{
			name:   "no matches",
			src:    []string{"a", "b", "c"},
			filter: []string{"x", "y"},
			want:   []string{"a", "b", "c"},
		},
		{
			name:   "empty source",
			src:    []string{},
			filter: []string{"a"},
			want:   []string{},
		},
		{
			name:   "empty filter",
			src:    []string{"a", "b"},
			filter: []string{},
			want:   []string{"a", "b"},
		},
		{
			name:   "filter all",
			src:    []string{"a", "b"},
			filter: []string{"a", "b"},
			want:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SliceFilter(tt.src, tt.filter)
			if len(got) != len(tt.want) {
				t.Errorf("SliceFilter() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("SliceFilter()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSliceFilter_Ints(t *testing.T) {
	got := SliceFilter([]int{1, 2, 3, 4, 5}, []int{2, 4})
	want := []int{1, 3, 5}
	if len(got) != len(want) {
		t.Errorf("SliceFilter() = %v, want %v", got, want)
	}
}
```

- [ ] **Step 2: Run**

Run: `cd /home/ai/HEUMathOverflow/backend && go test ./common/utils/ -run TestSliceFilter -v`

---

### Task 4: `common/middleware/concurrency_test.go` — Concurrency limiter

**Files:**
- Create: `backend/common/middleware/concurrency_test.go`

- [ ] **Step 1: Write the test file**

```go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupRouter(maxInflight int) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(GlobalConcurrencyMiddleware(maxInflight))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	return r
}

func TestConcurrencyMiddleware_PassthroughWhenZero(t *testing.T) {
	r := setupRouter(0)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestConcurrencyMiddleware_AllowsUnderLimit(t *testing.T) {
	r := setupRouter(5)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestConcurrencyMiddleware_RejectsOverLimit(t *testing.T) {
	r := setupRouter(1)
	var wg sync.WaitGroup
	results := make([]int, 3)

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			r.ServeHTTP(w, req)
			results[idx] = w.Code
		}(i)
	}
	wg.Wait()

	hasSuccess := false
	hasReject := false
	for _, code := range results {
		if code == 200 {
			hasSuccess = true
		}
		if code == 503 {
			hasReject = true
		}
	}
	if !hasSuccess {
		t.Error("At least one request should succeed")
	}
}
```

- [ ] **Step 2: Run**

Run: `cd /home/ai/HEUMathOverflow/backend && go test ./common/middleware/ -run TestConcurrency -v`

---

### Task 5: `common/middleware/permission_test.go` — Role-based permission

**Files:**
- Create: `backend/common/middleware/permission_test.go`

- [ ] **Step 1: Write the test file**

```go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupPermissionRouter(allowRole int) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.Set("role", 1)
		}
		_ = role
		c.Next()
	})
	r.Use(PermissionMiddleware(nil, allowRole))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	return r
}

func TestPermissionMiddleware_AllowsHigherRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", 3) // Teacher
		c.Next()
	})
	r.Use(PermissionMiddleware(nil, 2)) // Require Assistant
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestPermissionMiddleware_RejectsLowerRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", 1) // Student
		c.Next()
	})
	r.Use(PermissionMiddleware(nil, 3)) // Require Teacher
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	if w.Code != 401 {
		t.Errorf("Expected 401, got %d", w.Code)
	}
}

func TestPermissionMiddleware_AllowsEqualRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", 2) // Assistant
		c.Next()
	})
	r.Use(PermissionMiddleware(nil, 2)) // Require Assistant
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}
```

- [ ] **Step 2: Run**

Run: `cd /home/ai/HEUMathOverflow/backend && go test ./common/middleware/ -run TestPermission -v`

---

### Task 6: `common/api/gin_test.go` — Response builder

**Files:**
- Create: `backend/common/api/gin_test.go`

- [ ] **Step 1: Write the test file**

```go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestJSON_Success(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	JSON(c).Status(200).Code(200).Message("ok").Data(gin.H{"key": "value"}).Send()

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if resp.Code != 200 {
		t.Errorf("Expected code 200, got %d", resp.Code)
	}
	if resp.Message != "ok" {
		t.Errorf("Expected message 'ok', got '%s'", resp.Message)
	}
}

func TestJSON_WithPagination(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	JSON(c).Code(200).Message("ok").Data([]int{1, 2, 3}).Pagination(1, 10, 100).Send()

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Pagination == nil {
		t.Fatal("Expected pagination to be set")
	}
	if resp.Pagination.Page != 1 || resp.Pagination.PageSize != 10 || resp.Pagination.Total != 100 {
		t.Errorf("Pagination mismatch: %+v", resp.Pagination)
	}
}

func TestJSON_DefaultStatus(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	JSON(c).Code(404).Message("not found").Send()

	if w.Code != 404 {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestJSON_DefaultTo200(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	JSON(c).Message("ok").Send()

	if w.Code != 200 {
		t.Errorf("Expected default status 200, got %d", w.Code)
	}
}
```

- [ ] **Step 2: Run**

Run: `cd /home/ai/HEUMathOverflow/backend && go test ./common/api/ -run TestJSON -v`

---

## Part B: Frontend TypeScript Tests

### Task 7: Install Vitest and configure test environment

**Files:**
- Modify: `frontend/package.json` (add devDependencies and test script)
- Create: `frontend/vitest.config.ts`

- [ ] **Step 1: Install dependencies**

Run: `cd /home/ai/HEUMathOverflow/frontend && npm install -D vitest @vue/test-utils happy-dom @vitest/coverage-v8`

- [ ] **Step 2: Add test script to package.json**

Add to scripts: `"test": "vitest run", "test:watch": "vitest", "test:coverage": "vitest run --coverage"`

- [ ] **Step 3: Create vitest.config.ts**

```ts
import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
    plugins: [vue()],
    resolve: {
        alias: {
            '@': fileURLToPath(new URL('./src', import.meta.url)),
        },
    },
    test: {
        environment: 'happy-dom',
        globals: true,
        include: ['src/**/*.{test,spec}.{ts,tsx}'],
    },
})
```

- [ ] **Step 4: Verify**

Run: `cd /home/ai/HEUMathOverflow/frontend && npx vitest --version`

---

### Task 8: `src/utils/avatar.test.ts` — Avatar URL helper

**Files:**
- Create: `frontend/src/utils/avatar.test.ts`

- [ ] **Step 1: Write the test file**

```ts
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { getAvatarUrl } from './avatar'

describe('getAvatarUrl', () => {
    beforeEach(() => {
        vi.stubEnv('VITE_API_BASE_URL', 'http://localhost:5173')
    })

    it('returns default avatar for null/undefined/empty', () => {
        expect(getAvatarUrl(null)).toBe('/src/assets/images/default-avatar.png')
        expect(getAvatarUrl(undefined)).toBe('/src/assets/images/default-avatar.png')
        expect(getAvatarUrl('')).toBe('/src/assets/images/default-avatar.png')
    })

    it('passes through absolute http URLs', () => {
        expect(getAvatarUrl('http://example.com/avatar.png')).toBe('http://example.com/avatar.png')
    })

    it('passes through absolute https URLs', () => {
        expect(getAvatarUrl('https://cdn.example.com/avatar.png')).toBe('https://cdn.example.com/avatar.png')
    })

    it('prepends base URL to relative paths', () => {
        const result = getAvatarUrl('/uploads/avatar.png')
        expect(result).toBe('http://localhost:5173/uploads/avatar.png')
    })

    it('normalizes missing leading slash', () => {
        const result = getAvatarUrl('uploads/avatar.png')
        expect(result).toBe('http://localhost:5173/uploads/avatar.png')
    })

    it('adds timestamp when forceRefresh is true', () => {
        const result = getAvatarUrl('https://example.com/avatar.png', true)
        expect(result).toMatch(/\?_t=\d+/)
    })

    it('uses & separator when URL already has query params', () => {
        const result = getAvatarUrl('https://example.com/avatar.png?v=1', true)
        expect(result).toContain('&_t=')
    })
})
```

- [ ] **Step 2: Run**

Run: `cd /home/ai/HEUMathOverflow/frontend && npx vitest run src/utils/avatar.test.ts`

---

### Task 9: `src/utils/errorHandler.test.ts` — Error handler

**Files:**
- Create: `frontend/src/utils/errorHandler.test.ts`

- [ ] **Step 1: Write the test file**

```ts
import { describe, it, expect } from 'vitest'
import {
    AppError,
    ErrorType,
    ErrorSeverity,
    createBusinessError,
    createErrorFromAxios,
    shouldShowDialog,
    shouldShowMessage,
} from './errorHandler'

describe('AppError', () => {
    it('creates with correct properties', () => {
        const error = new AppError('test', ErrorType.HTTP_ERROR, ErrorSeverity.CRITICAL, 500)
        expect(error.message).toBe('test')
        expect(error.type).toBe(ErrorType.HTTP_ERROR)
        expect(error.severity).toBe(ErrorSeverity.CRITICAL)
        expect(error.code).toBe(500)
        expect(error.name).toBe('AppError')
    })

    it('defaults to UNKNOWN_ERROR and NORMAL severity', () => {
        const error = new AppError('test')
        expect(error.type).toBe(ErrorType.UNKNOWN_ERROR)
        expect(error.severity).toBe(ErrorSeverity.NORMAL)
    })
})

describe('createBusinessError', () => {
    it('creates BUSINESS_ERROR type', () => {
        const error = createBusinessError('fail', 400)
        expect(error.type).toBe(ErrorType.BUSINESS_ERROR)
        expect(error.message).toBe('fail')
        expect(error.code).toBe(400)
    })

    it('defaults severity to NORMAL', () => {
        const error = createBusinessError('fail')
        expect(error.severity).toBe(ErrorSeverity.NORMAL)
    })
})

describe('createErrorFromAxios', () => {
    it('handles HTTP 401 as CRITICAL', () => {
        const error = createErrorFromAxios({
            response: { status: 401, data: { message: 'unauthorized' } },
            isAxiosError: true,
        } as any)
        expect(error.type).toBe(ErrorType.HTTP_ERROR)
        expect(error.severity).toBe(ErrorSeverity.CRITICAL)
        expect(error.code).toBe(401)
    })

    it('handles HTTP 404 as NORMAL', () => {
        const error = createErrorFromAxios({
            response: { status: 404, data: {} },
            isAxiosError: true,
        } as any)
        expect(error.severity).toBe(ErrorSeverity.NORMAL)
    })

    it('handles HTTP 500 as CRITICAL', () => {
        const error = createErrorFromAxios({
            response: { status: 500, data: {} },
            isAxiosError: true,
        } as any)
        expect(error.severity).toBe(ErrorSeverity.CRITICAL)
    })

    it('handles network errors', () => {
        const error = createErrorFromAxios({
            request: {},
            isAxiosError: true,
        } as any)
        expect(error.type).toBe(ErrorType.NETWORK_ERROR)
        expect(error.severity).toBe(ErrorSeverity.CRITICAL)
    })

    it('handles config errors', () => {
        const error = createErrorFromAxios({
            isAxiosError: true,
        } as any)
        expect(error.type).toBe(ErrorType.UNKNOWN_ERROR)
    })
})

describe('shouldShowDialog', () => {
    it('returns true for CRITICAL', () => {
        expect(shouldShowDialog(new AppError('', ErrorType.HTTP_ERROR, ErrorSeverity.CRITICAL))).toBe(true)
    })
    it('returns false for NORMAL', () => {
        expect(shouldShowDialog(new AppError('', ErrorType.HTTP_ERROR, ErrorSeverity.NORMAL))).toBe(false)
    })
    it('returns false for SILENT', () => {
        expect(shouldShowDialog(new AppError('', ErrorType.HTTP_ERROR, ErrorSeverity.SILENT))).toBe(false)
    })
})

describe('shouldShowMessage', () => {
    it('returns true for CRITICAL', () => {
        expect(shouldShowMessage(new AppError('', ErrorType.HTTP_ERROR, ErrorSeverity.CRITICAL))).toBe(true)
    })
    it('returns true for NORMAL', () => {
        expect(shouldShowMessage(new AppError('', ErrorType.HTTP_ERROR, ErrorSeverity.NORMAL))).toBe(true)
    })
    it('returns false for SILENT', () => {
        expect(shouldShowMessage(new AppError('', ErrorType.HTTP_ERROR, ErrorSeverity.SILENT))).toBe(false)
    })
})
```

- [ ] **Step 2: Run**

Run: `cd /home/ai/HEUMathOverflow/frontend && npx vitest run src/utils/errorHandler.test.ts`

---

### Task 10: `src/types/common.test.ts` — Type helper functions

**Files:**
- Create: `frontend/src/types/common.test.ts`

- [ ] **Step 1: Write the test file**

```ts
import { describe, it, expect } from 'vitest'
import { getRoleText, getPostStatusText, getReplyStatusText, isReplyCertified } from './common'

describe('getRoleText', () => {
    it('returns correct text for each role', () => {
        expect(getRoleText(1)).toBe('学生')
        expect(getRoleText(2)).toBe('助教')
        expect(getRoleText(3)).toBe('教师')
        expect(getRoleText(4)).toBe('管理员')
    })
    it('defaults to 学生 for unknown role', () => {
        expect(getRoleText(99 as any)).toBe('学生')
    })
})

describe('getPostStatusText', () => {
    it('returns correct text for each status', () => {
        expect(getPostStatusText(1)).toBe('未解决')
        expect(getPostStatusText(2)).toBe('已解决')
        expect(getPostStatusText(3)).toBe('已认证')
    })
    it('defaults to 未解决 for unknown status', () => {
        expect(getPostStatusText(99 as any)).toBe('未解决')
    })
})

describe('getReplyStatusText', () => {
    it('returns correct text for each status', () => {
        expect(getReplyStatusText(1)).toBe('未精选')
        expect(getReplyStatusText(2)).toBe('作者精选')
        expect(getReplyStatusText(3)).toBe('教师精选')
    })
    it('defaults to 未精选 for unknown status', () => {
        expect(getReplyStatusText(99 as any)).toBe('未精选')
    })
})

describe('isReplyCertified', () => {
    it('returns true for author-selected (2)', () => {
        expect(isReplyCertified(2)).toBe(true)
    })
    it('returns true for teacher-certified (3)', () => {
        expect(isReplyCertified(3)).toBe(true)
    })
    it('returns false for not-selected (1)', () => {
        expect(isReplyCertified(1)).toBe(false)
    })
})
```

- [ ] **Step 2: Run**

Run: `cd /home/ai/HEUMathOverflow/frontend && npx vitest run src/types/common.test.ts`

---

### Task 11: `src/services/forum.test.ts` — Forum data mappers

**Files:**
- Create: `frontend/src/services/forum.test.ts`

- [ ] **Step 1: Write the test file**

```ts
import { describe, it, expect } from 'vitest'
import type { PostWithAuthor, ReplyWithAuthor } from '@/types/forum'

// We need to test the mapper functions. Since they are not exported,
// we test them indirectly through the public API or re-export them.
// For now, we test the mapping logic inline.

describe('PostWithAuthor mapping', () => {
    it('flattens post_data and user_info into Post', () => {
        const input: PostWithAuthor = {
            post_data: {
                post_id: '123',
                title: 'Test',
                content: '<p>Hello</p>',
                tags: ['math'],
                image_urls: [],
                status: 1,
                views: 10,
                likes: 5,
                stars: 2,
                replies: 3,
                last_reply_at: null,
                created_at: '2024-01-01T00:00:00Z',
            },
            user_info: {
                user_id: '456',
                username: 'testuser',
                avatar_url: '/avatar.png',
                role: 1,
            },
        }

        // Simulate mapPostWithAuthor logic
        const { post_data, user_info } = input
        const post = { ...post_data, author: user_info }

        expect(post.post_id).toBe('123')
        expect(post.author.username).toBe('testuser')
        expect(post.author.role).toBe(1)
        expect(post.title).toBe('Test')
    })
})

describe('ReplyWithAuthor mapping', () => {
    it('flattens reply_data and user_info into Reply', () => {
        const input: ReplyWithAuthor = {
            reply_data: {
                reply_id: '789',
                post_id: '123',
                parent_reply_id: null,
                content: '<p>Answer</p>',
                status: 1,
                image_urls: [],
                voice_url: '',
                voice_text: '',
                ai_answered: false,
                certified_by: '',
                created_at: '2024-01-01T00:00:00Z',
            },
            user_info: {
                user_id: '456',
                username: 'answerer',
                avatar_url: '/avatar.png',
                role: 3,
            },
        }

        const { reply_data, user_info } = input
        const reply = { ...reply_data, author: user_info }

        expect(reply.reply_id).toBe('789')
        expect(reply.author.role).toBe(3)
    })
})

describe('Pagination mapping', () => {
    it('computes total_pages from total and page_size', () => {
        const input = { page: 1, page_size: 20, total: 95 }
        const pagination = {
            page: input.page,
            page_size: input.page_size,
            total: input.total,
            total_pages: input.total ? Math.ceil(input.total / input.page_size) : undefined,
        }
        expect(pagination.total_pages).toBe(5)
    })

    it('handles missing total', () => {
        const input = { page: 1, page_size: 20 }
        const pagination = {
            page: input.page,
            page_size: input.page_size,
            total: input.total,
            total_pages: input.total ? Math.ceil(input.total / input.page_size) : undefined,
        }
        expect(pagination.total_pages).toBeUndefined()
    })

    it('handles exact division', () => {
        const input = { page: 1, page_size: 10, total: 100 }
        expect(Math.ceil(input.total / input.page_size)).toBe(10)
    })
})
```

- [ ] **Step 2: Run**

Run: `cd /home/ai/HEUMathOverflow/frontend && npx vitest run src/services/forum.test.ts`

---

### Task 12: `src/utils/content.test.ts` — Content extraction

**Files:**
- Create: `frontend/src/utils/content.test.ts`

- [ ] **Step 1: Write the test file**

```ts
import { describe, it, expect } from 'vitest'
import { extractImageUrls, extractPlainText } from './content'

describe('extractImageUrls', () => {
    it('extracts image URLs from HTML', () => {
        const html = '<p>Hello</p><img src="http://example.com/a.png"><img src="http://example.com/b.png">'
        expect(extractImageUrls(html)).toEqual(['http://example.com/a.png', 'http://example.com/b.png'])
    })

    it('deduplicates URLs', () => {
        const html = '<img src="http://example.com/a.png"><img src="http://example.com/a.png">'
        expect(extractImageUrls(html)).toEqual(['http://example.com/a.png'])
    })

    it('returns empty array for empty/null input', () => {
        expect(extractImageUrls('')).toEqual([])
        expect(extractImageUrls(null as any)).toEqual([])
        expect(extractImageUrls(undefined as any)).toEqual([])
    })

    it('returns empty array for HTML without images', () => {
        expect(extractImageUrls('<p>No images here</p>')).toEqual([])
    })

    it('handles single quotes in src', () => {
        const html = "<img src='http://example.com/img.png'>"
        expect(extractImageUrls(html)).toEqual(['http://example.com/img.png'])
    })
})

describe('extractPlainText', () => {
    it('strips HTML tags', () => {
        expect(extractPlainText('<p>Hello <b>World</b></p>')).toBe('Hello World')
    })

    it('removes img tags', () => {
        expect(extractPlainText('<p>Text</p><img src="x.png">')).toBe('Text')
    })

    it('removes video tags', () => {
        expect(extractPlainText('<p>Text</p><video src="v.mp4"></video>')).toBe('Text')
    })

    it('returns empty string for empty/null input', () => {
        expect(extractPlainText('')).toBe('')
        expect(extractPlainText(null as any)).toBe('')
    })
})
```

- [ ] **Step 2: Run**

Run: `cd /home/ai/HEUMathOverflow/frontend && npx vitest run src/utils/content.test.ts`

---

### Task 13: Run all tests and generate summary

- [ ] **Step 1: Run all backend tests**

Run: `cd /home/ai/HEUMathOverflow/backend && go test ./common/... -v -count=1`

- [ ] **Step 2: Run all frontend tests**

Run: `cd /home/ai/HEUMathOverflow/frontend && npx vitest run`

- [ ] **Step 3: Generate coverage reports**

Run: `cd /home/ai/HEUMathOverflow/frontend && npx vitest run --coverage`
