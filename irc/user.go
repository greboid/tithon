package irc

type User struct {
	nickname string
}

func NewUser(nickname string) *User {
	return &User{nickname}
}

func (u *User) GetNickListDisplay() string {
	return u.nickname
}
