package auth

import (
	"crypto/md5"
	"encoding/hex"
	"time"

	"github.com/knchan0x/belle-maison/internal/cache"
)

var (
	// default
	adminId = "admin"
	adminPw = "admin"
)

func SetAdmin(id, pw string) {
	adminId = id
	adminPw = pw
}

var cookieName = "_cookie_"

func GetCookieName() string {
	return cookieName
}

func SetCookieName(name string) {
	cookieName = name
}

const (
	salt     = "belle-masion"
	tokenKey = "token"
)

var sessionDB = cache.New("InMemory")

// VerifyToken check is token provided exist in cache
// TODO: return user info
func VerifyToken(token string) bool {
	if t, ok := sessionDB.Get(tokenKey); !ok || token != t.(string) {
		return false
	}
	return true
}

func VerifyUser(username, password string) (string, bool) {
	if pw, exists := getUserPassword(username); !exists || pw != password {
		return "", false
	}

	// generate token
	md5 := md5.New()
	md5.Write([]byte(time.Now().Format("2006-01-02 15:04:05") + salt))
	token := hex.EncodeToString(md5.Sum(nil))

	sessionDB.Add(tokenKey, token, time.Hour)
	return token, true
}

func getUserPassword(username string) (password string, exists bool) {
	if username != adminId {
		return adminPw, true
	}
	return "", false
}
