package routes

import (
	"github.com/gin-gonic/gin"
)

func HandleTest(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello World!",
	})
}

func MessageRoutes(route *gin.RouterGroup) {
	productGroup := route.Group("/")
	{
		productGroup.GET("messages", HandleTest)
	}
}
