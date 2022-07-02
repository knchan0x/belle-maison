package auth

import (
	"crypto/md5"
	"encoding/hex"
	"time"

	"github.com/knchan0x/belle-maison/cmd/web/user"
	"github.com/knchan0x/belle-maison/internal/cache"
)

var cookieName = "_cookie_"

func GetCookieName() string {
	return cookieName
}

func SetCookieName(name string) {
	cookieName = name
}

const (
	salt = "belle-masion"
)

var sessionDB = cache.New("InMemory")

// VerifyToken check is token provided exist in cache
func VerifyToken(token string) user.User {
	u, ok := sessionDB.Get(token)
	if !ok {
		return user.User{Username: "Guest", Role: user.Guest}
	}
	return u.(user.User)
}

func VerifyUser(username, password string) (string, bool) {
	user, pw, exists := user.GetUser(username)
	if !exists || pw != password {
		return "", false
	}

	// generate token
	md5 := md5.New()
	md5.Write([]byte(time.Now().Format("2006-01-02 15:04:05") + salt))
	token := hex.EncodeToString(md5.Sum(nil))

	sessionDB.Add(token, *user, time.Hour)
	return token, true
}
