package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AllowCrossOrigin middleware handles CORS issues
// when debuging Vue app in http://localhost:3000
func AllowCrossOrigin(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "http://localhost:3000")
	ctx.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, PATCH, OPTIONS")
	ctx.Header("Access-Control-Allow-Headers", "Content-Type")
	ctx.Header("Access-Control-Max-Age", "86400")

	if ctx.Request.Method == http.MethodOptions {
		ctx.Status(http.StatusOK)
		return
	}
	ctx.Next()
}
