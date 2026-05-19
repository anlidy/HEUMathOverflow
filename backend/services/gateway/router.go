package main

import (
	"MathOverflow/common/middleware"
	"net/http/httputil"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// 用户服务路由
func SetupUserRouter(rdb *redis.Client, r *gin.Engine, userProxy *httputil.ReverseProxy) *gin.Engine {
	api := r.Group("/api/v1/user")

	// 无需验证的路由
	api.POST("/login", proxyTo(userProxy, "", "/user/login"))
	api.POST("/register", proxyTo(userProxy, "", "/user/register"))
	api.GET("/avatar/*filepath", proxyTo(userProxy, "/api/v1/user", "/user"))

	auth := r.Group("/api/v1/user")
	auth.Use(middleware.AuthMiddleware(rdb))
	auth.POST("/avatar", proxyTo(userProxy, "", "/user/avatar"))
	auth.PATCH("/profile", proxyTo(userProxy, "", "/user/profile"))
	auth.PATCH("/password", proxyTo(userProxy, "", "/user/password"))
	auth.POST("/logout", proxyTo(userProxy, "", "/user/logout"))
	auth.DELETE("/account", proxyTo(userProxy, "", "/user/account"))

	return r
}

// 论坛服务路由
func SetupForumRouter(rdb *redis.Client, r *gin.Engine, forumProxy *httputil.ReverseProxy) *gin.Engine {
	api := r.Group("/api/v1/forum")

	// 无需验证的路由
	api.GET("/file/*filename", proxyTo(forumProxy, "/api/v1/forum", "/forum"))

	auth := r.Group("/api/v1/forum")
	auth.Use(middleware.AuthMiddleware(rdb))
	auth.POST("/upload", proxyTo(forumProxy, "", "/forum/upload"))
	auth.POST("/posts", proxyTo(forumProxy, "", "/forum/posts"))
	auth.GET("/posts", proxyTo(forumProxy, "", "/forum/posts"))
	auth.Any("/posts/*path", proxyTo(forumProxy, "/api/v1/forum", "/forum"))
	auth.POST("/replies", proxyTo(forumProxy, "", "/forum/replies"))
	auth.Any("/replies/*path", proxyTo(forumProxy, "/api/v1/forum", "/forum"))
	auth.POST("/search", proxyTo(forumProxy, "", "/forum/search"))

	return r
}

func proxyTo(proxy *httputil.ReverseProxy, sourcePrefix string, target string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if sourcePrefix == "" {
			c.Request.URL.Path = target
		} else {
			c.Request.URL.Path = target + strings.TrimPrefix(c.Request.URL.Path, sourcePrefix)
		}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
