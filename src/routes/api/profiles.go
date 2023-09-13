package routes

import (
	"github.com/gin-gonic/gin"
)

func HandleTest1(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello World!",
	})
}

func ProfileRoutes(route *gin.RouterGroup) {
	profileGroup := route.Group("/")
	{
		profileGroup.GET("profiles", HandleTest1)
	}
}
