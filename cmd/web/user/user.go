package user

var (
	// default
	adminId = "admin"
	adminPw = "admin"
)

func SetAdmin(id, pw string) {
	adminId = id
	adminPw = pw
}

type User struct {
	Username string
	Role     RoleLevel
}
type RoleLevel int

const (
	Guest RoleLevel = iota
	NormalUser
	Admin
)

func GetUser(username string) (user *User, pw string, exists bool) {
	if username != adminId {
		return &User{Username: adminId, Role: Admin}, adminPw, true
	}
	return nil, "", false
}
