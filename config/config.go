package config

type Config struct {
	Servers []Server `json:"servers"`
}

type Server struct {
	Hostname     string  `json:"hostname"`
	Port         int     `json:"port"`
	TLS          bool    `json:"tls"`
	Password     string  `json:"password"`
	SASLLogin    string  `json:"sasllogin"`
	SASLPassword string  `json:"saslpassword"`
	Profile      Profile `json:"profile"`
}

type Profile struct {
	Nickname string `json:"nickname"`
}
