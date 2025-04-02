package irc

type Profile struct {
	nickname string
}

func NewProfile(nickname string) *Profile {
	return &Profile{
		nickname: nickname,
	}
}

func (p *Profile) GetNickname() string {
	return p.nickname
}
