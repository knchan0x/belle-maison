package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AllowCrossOrigin returns gin middleware handles CORS issues
// for domain provided i.e. http://localhost:3000
// when debuging Vue app
func AllowCrossOrigin(domain string) func(*gin.Context) {
	return func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", domain)
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, PATCH, OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type")
		ctx.Header("Access-Control-Max-Age", "86400")

		if ctx.Request.Method == http.MethodOptions {
			ctx.Status(http.StatusOK)
			return
		}
		ctx.Next()
	}
}
