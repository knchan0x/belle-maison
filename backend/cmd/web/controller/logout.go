package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/knchan0x/belle-maison/backend/cmd/web/auth"
)

// logout handler, remove token and jump to login page
func Logout(jumpTo string) func(*gin.Context) {
	return func(ctx *gin.Context) {
		t, err := ctx.Cookie(auth.GetCookieName())
		if err == nil {
			auth.RemoveToken(t)
		}
		ctx.SetCookie(auth.GetCookieName(), "", -1, "/", "", false, true) // secure flag causes only https allowed to set cookie
		ctx.Redirect(http.StatusFound, jumpTo)
	}
}
