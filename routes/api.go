package routes

import (
	"github.com/gin-gonic/gin"
)

func RouteApi(router *gin.RouterGroup) {
	router.GET("/test", func(context *gin.Context) {
		context.JSON(200, gin.H{
			"status": "ok",
		})
	})

	router.GET("/users", func(ctx *gin.Context) {
	})
}
