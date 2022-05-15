package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/knchan0x/belle-maison/cmd/web/auth"
)

type AuthMode string

const (
	AuthMode_Redirect     AuthMode = "Redirect"
	AuthMode_Unauthorized AuthMode = "Unauthorized"
)

var simpleAuthSuspended = false

func ActivateSimpleAuth(isActivate bool) {
	simpleAuthSuspended = !isActivate
}

// SimpleAuth returns gin middleware with mode specified.
// This middleware will check is the user permit to access.
// It uses cookie to store the token.
//
// - AuthMode_Redirect = redirect to login page, will jump to "/" if not provided
//
// - AuthMode_Unauthorized = return JSON with 401 unauthorized
func SimpleAuth(mode AuthMode, jumpTo ...string) func(ctx *gin.Context) {
	if !simpleAuthSuspended {
		if mode == AuthMode_Redirect {
			return func(ctx *gin.Context) {
				t, err := ctx.Cookie(auth.GetCookieName())
				if err != nil || !auth.VerifyToken(t) {
					if len(jumpTo) != 1 {
						ctx.Redirect(http.StatusFound, "/")
					} else {
						ctx.Redirect(http.StatusFound, jumpTo[0])
					}
					return
				}
				ctx.Next()
			}
		}
		if mode == AuthMode_Unauthorized {
			return func(ctx *gin.Context) {
				t, err := ctx.Cookie(auth.GetCookieName())
				if err != nil || !auth.VerifyToken(t) {
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
