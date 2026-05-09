package main

import (
	"MathOverflow/internal/common/middleware"
	"net/http/httputil"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// 用户服务路由
func SetupUserRouter(rdb *redis.Client, r *gin.Engine, userProxy *httputil.ReverseProxy) *gin.Engine {
	api := r.Group("/api/v1/user")

	// 无需验证的路由
	api.POST("/login", proxyTo(userProxy, "/user/login"))
	api.POST("/register", proxyTo(userProxy, "/user/register"))
	api.GET("/avatar/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		c.Request.URL.Path = "/user/avatar" + filepath
		userProxy.ServeHTTP(c.Writer, c.Request)
	})

	// 需要验证的路由
	api.Use(middleware.AuthMiddleware(rdb))
	api.Any("/*path", func(c *gin.Context) {
		userProxy.ServeHTTP(c.Writer, c.Request)
	})

	return r
}

// 论坛服务路由
func SetupForumRouter(rdb *redis.Client, r *gin.Engine, forumProxy *httputil.ReverseProxy) *gin.Engine {
	api := r.Group("/api/v1/forum")

	// 无需验证的路由
	api.GET("/file/*filename", func(c *gin.Context) {
		filename := c.Param("filename")
		c.Request.URL.Path = "/forum/file" + filename
		forumProxy.ServeHTTP(c.Writer, c.Request)
	})

	// 需要验证的路由
	api.Use(middleware.AuthMiddleware(rdb))
	api.Any("/*path", func(c *gin.Context) {
		forumProxy.ServeHTTP(c.Writer, c.Request)
	})

	return r
}

// 论坛服务路由
func proxyTo(proxy *httputil.ReverseProxy, targetPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.URL.Path = targetPath
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
