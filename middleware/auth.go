package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

func Auth(exclude []string) gin.HandlerFunc {
	return func(context *gin.Context) {
		t := time.Now()
		fmt.Printf("auth time: %s\n", t)
		context.Next()
	}
}