package routes

import (
	"github.com/gin-gonic/gin"
)

func HandleLogin(c *gin.Context) {}

func HandleRegister(c *gin.Context) {}

func HandleRevokeToken(c *gin.Context) {}

func HandleRefreshToken(c *gin.Context) {}

func AuthenticationRoutes(route *gin.RouterGroup) {
	authGroup := route.Group("/")
	{
		authGroup.POST("login", HandleLogin)
		authGroup.POST("register", HandleRegister)
		authGroup.POST("revoke", HandleRevokeToken)
		authGroup.POST("refresh_token", HandleRefreshToken)
	}
}
