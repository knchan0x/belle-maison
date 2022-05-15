package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/knchan0x/belle-maison/internal/cache"
)

type AuthMode string

const (
	cookie_name = "_cookie_belle_masion"

	AuthMode_Redirect     AuthMode = "Redirect"
	AuthMode_Unauthorized AuthMode = "Unauthorized"

	urlPath_root    = "/bellemasion"
	urlPrefix_login = "/login"
	urlPath_login   = urlPath_root + urlPrefix_login
)

var simpleAuthSuspended = false

func ActivateSimpleAuth(isActivate bool) {
	simpleAuthSuspended = !isActivate
}

// SimpleAuth returns gin middleware with mode specified.
// This middleware will check is the user permit to access
//
// - AuthMode_Redirect = redirect to login page
// - AuthMode_Unauthorized = return JSON with 401 unauthorized
func SimpleAuth(mode AuthMode) func(ctx *gin.Context) {
	if !simpleAuthSuspended {
		if mode == AuthMode_Redirect {
			return func(ctx *gin.Context) {
				t, err := ctx.Cookie(cookie_name)
				token, ok := cache.Get("token")
				if err != nil || !ok || t != token.(string) {
					ctx.Redirect(http.StatusFound, urlPath_login)
					return
				}
				ctx.Next()
			}
		}
		if mode == AuthMode_Unauthorized {
			return func(ctx *gin.Context) {
				t, err := ctx.Cookie(cookie_name)
				token, ok := cache.Get("token")
				if err != nil || !ok || t != token.(string) {
					ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}
				ctx.Next()
			}
		}
	}

	return func(ctx *gin.Context) {
		ctx.Next() // by pass
	}
}
