package irc

type User struct {
	nickname string
	modes    string
}

func NewUser(nickname string, modes string) *User {
	return &User{nickname: nickname, modes: modes}
}

func (u *User) GetNickListDisplay() string {
	return u.nickname
}

func (u *User) GetNickListModes() string {
	return u.modes
}
