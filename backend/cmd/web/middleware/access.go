package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/knchan0x/belle-maison/backend/cmd/web/auth"
)

type AuthMode string

const (
	AuthMode_Redirect     AuthMode = "Redirect"
	AuthMode_Unauthorized AuthMode = "Unauthorized"
)

type User struct {
	Username string
	Role     string
}

type RoleLevel int

const (
	Guest RoleLevel = iota
	NormalUser
	Admin
)

var rolePermitSuspended = false

func ActivateRolePermit(isActivate bool) {
	rolePermitSuspended = !isActivate
}

// AccessControl returns gin middleware with mode specified.
// This middleware will check is the user permit to access.
// It uses cookie to store the token.
// It will also save the user info in *gin.Content with key "User".
//
// - AuthMode_Redirect = redirect to login page, will jump to "/" if not provided
//
// - AuthMode_Unauthorized = return JSON with 401 unauthorized
func AccessControl(level RoleLevel, mode AuthMode, jumpTo ...string) func(ctx *gin.Context) {
	if !rolePermitSuspended {
		if mode == AuthMode_Redirect {
			return func(ctx *gin.Context) {
				t, err := ctx.Cookie(auth.GetCookieName())
				currentUser := auth.VerifyToken(t)
				if err != nil || RoleLevel(currentUser.Role) < level {
					if len(jumpTo) != 1 {
						ctx.Redirect(http.StatusFound, "/")
					} else {
						ctx.Redirect(http.StatusFound, jumpTo[0])
					}
					return
				}
				ctx.Set("User", currentUser)
				ctx.Next()
			}
		}
		if mode == AuthMode_Unauthorized {
			return func(ctx *gin.Context) {
				t, err := ctx.Cookie(auth.GetCookieName())
				currentUser := auth.VerifyToken(t)
				if err != nil || RoleLevel(currentUser.Role) < level {
					ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}
				ctx.Set("User", currentUser)
				ctx.Next()
			}
		}
	}

	return func(ctx *gin.Context) {
		ctx.Next() // by pass
	}
}
