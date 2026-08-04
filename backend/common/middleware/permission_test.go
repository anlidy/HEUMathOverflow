package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupPermRouter(allowRole int) *gin.Engine {
	router := gin.New()
	router.Use(PermissionMiddleware(nil, allowRole))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	return router
}

func TestPermissionMiddleware_AllowsHigherRole(t *testing.T) {
	router := setupPermRouter(2)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// Set role=3 (Teacher) via Gin context
	router.Use(func(c *gin.Context) {
		c.Set("role", 3)
	})
	// Need a fresh router since Use must be before routes
	router = gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", 3)
	})
	router.Use(PermissionMiddleware(nil, 2))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for higher role, got %d", w.Code)
	}
}

func TestPermissionMiddleware_RejectsLowerRole(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", 1)
	})
	router.Use(PermissionMiddleware(nil, 3))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for lower role, got %d", w.Code)
	}
}

func TestPermissionMiddleware_AllowsEqualRole(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", 2)
	})
	router.Use(PermissionMiddleware(nil, 2))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for equal role, got %d", w.Code)
	}
}
