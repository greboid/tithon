package events

import "time"

const v3TimestampFormat = "2006-01-02T15:04:05.000Z"

type IRCTime struct {
	time.Time
}

type Profile struct {
	Nick     string `yaml:"nick" json:"nick"`
	User     string `yaml:"user,omitempty" json:"user,omitempty"`
	Realname string `yaml:"realname,omitempty" json:"realname,omitempty"`
}
type Server struct {
	ID           string     `yaml:"-" json:"id"`
	Server       string     `yaml:"server" json:"server"`
	TLS          bool       `yaml:"tls" json:"tls"`
	SaslMech     string     `yaml:"saslMech,omitempty" json:"saslMech,omitempty"`
	Saslusername string     `yaml:"saslUsername,omitempty" json:"saslUsername,omitempty"`
	Saslpassword string     `yaml:"saslPassword,omitempty" json:"saslPassword,omitempty"`
	Profile      *Profile   `yaml:"profile" json:"profile"`
	Channels     []*Channel `yaml:"-" json:"channels"`
}

type User struct {
	Nick     string `yaml:"-" json:"nick"`
	UserHost string `yaml:"-" json:"userhost"`
	Realname string `yaml:"-" json:"realname"`
}

type ChannelUser struct {
	User
	Modes string `yaml:"-" json:"modes"`
}

type ModeNoParam struct {
	Char rune `yaml:"-" json:"char"`
}

type ModeList struct {
	Char    rune     `yaml:"-" json:"char"`
	Entries []string `yaml:"-" json:"entries"`
}

type ModeParamSetUnset struct {
	Char  rune   `yaml:"-" json:"char"`
	Value string `yaml:"-" json:"value"`
}

type ModeParamSet struct {
	Char  rune   `yaml:"-" json:"char"`
	Value string `yaml:"-" json:"value"`
}

type Channel struct {
	Server             *Server             `yaml:"-" json:"server"`
	Name               string              `yaml:"-" json:"name"`
	Users              []*ChannelUser      `yaml:"-" json:"users"`
	Topic              string              `yaml:"-" json:"topic"`
	ModesList          []ModeList          `json:"modeslist"`
	ModesNoParam       []ModeNoParam       `json:"modesnoparam"`
	ModesParamSet      []ModeParamSet      `json:"modesparamset"`
	ModesParamSetUnset []ModeParamSetUnset `json:"modesparamsetunset"`
}

type ChannelMessage struct {
	Channel  *Channel `yaml:"-" json:"channel"`
	Message  string   `yaml:"message" json:"message"`
	IsNotice bool     `yaml:"-" json:"isNotice,omitempty"`
	IsAction bool     `yaml:"-" json:"isAction,omitempty"`
}

type DirectMessage struct {
	Server   *Server `yaml:"-" json:"server"`
	User     *User   `yaml:"source" json:"source"`
	Message  string  `yaml:"message" json:"message"`
	IsNotice bool    `yaml:"-" json:"isNotice,omitempty"`
	IsAction bool    `yaml:"-" json:"isAction,omitempty"`
}

type ServerMessage struct {
	Server   *Server `yaml:"-" json:"server"`
	Message  string  `yaml:"message" json:"message"`
	IsNotice bool    `yaml:"-" json:"isNotice,omitempty"`
	IsAction bool    `yaml:"-" json:"isAction,omitempty"`
}
