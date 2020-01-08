package routes

import "github.com/gin-gonic/gin"

func RouteWeb(router *gin.RouterGroup) {
	router.GET("/test", func(context *gin.Context) {
		context.JSON(200, gin.H{
			"status": "ok",
			"path":   "/test",
		})
	})
}
