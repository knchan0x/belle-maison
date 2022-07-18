package middleware

import (
	"github.com/gin-gonic/gin"
)

// AllowCrossOrigin returns gin middleware handles CORS issues
// for domain provided i.e. http://localhost:3000
// when debuging Vue app
func NoCache() func(*gin.Context) {
	return func(ctx *gin.Context) {
		ctx.Header("Cache-Control", "no-cache")
		ctx.Next()
	}
}
