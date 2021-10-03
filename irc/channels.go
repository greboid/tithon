package irc

type Channel struct {
	Subject        string           `yaml:"-"`
	Nicknames      []User           `yaml:"-"`
	RecentMessages []ChannelMessage `yaml:"-"`
	Name           string           `yaml:"name"`
}

type ChannelMessage struct {
	Source  User
	Time    int64
	Message string
}
