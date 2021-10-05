package irc

type Channel struct {
	Subject        string           `json:"-"`
	Nicknames      []User           `json:"-"`
	RecentMessages []ChannelMessage `json:"-"`
	Name           string           `json:"name"`
	Joined         bool             `json:"-"`
}

type ChannelMessage struct {
	Source  User
	Time    int64
	Message string
}
