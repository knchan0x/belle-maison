package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/knchan0x/belle-maison/backend/cmd/web/auth"
)

// login handler, jump to provided address if user is logged in.
// return "username and/or password incorrect." if invalid
// Insecure. Please set SSL.
func Login(jumpTo string) func(*gin.Context) {
	return func(ctx *gin.Context) {
		token, ok := auth.VerifyUser(ctx.PostForm("username"), ctx.PostForm("password"))

		if !ok {
			ctx.String(http.StatusUnauthorized, "username and/or password incorrect.")
			return
		}

		ctx.SetCookie(auth.GetCookieName(), token, 60*60, "/", "", false, true) // secure flag causes only https allowed to set cookie
		ctx.Redirect(http.StatusFound, jumpTo)
	}
}
