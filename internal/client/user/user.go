package user

type User struct {
	ID       int
	Password string
}

var CurrentUser *User
