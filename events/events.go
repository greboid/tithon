package events

type ConnectableServer struct {
	Server       string             `yaml:"server" json:"server"`
	TLS          bool               `yaml:"tls" json:"tls"`
	SaslMech     string             `yaml:"saslMech,omitempty" json:"saslMech,omitempty"`
	Saslusername string             `yaml:"saslUsername,omitempty" json:"saslUsername,omitempty"`
	Saslpassword string             `yaml:"saslPassword,omitempty" json:"saslPassword,omitempty"`
	Profile      ConnectableProfile `yaml:"profile" json:"profile"`
}

type ConnectableProfile struct {
	Nick     string `yaml:"nick" json:"nick"`
	User     string `yaml:"user,omitempty" json:"user,omitempty"`
	Realname string `yaml:"realname,omitempty" json:"realname,omitempty"`
}

type Channel struct {
	Name string `json:"name" yaml:"name"`
}

type Message struct {
	Source  string `json:"source" yaml:"source"`
	Target  string `json:"target" yaml:"target"`
	Message string `json:"message" yaml:"message"`
}

type ChannelMessage struct {
	Message
}

type DirectMessage struct {
	Message
}

type ServerMessage struct {
	Message
}
