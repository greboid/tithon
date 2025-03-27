package irc

import (
	uniqueid "github.com/albinj12/unique-id"
	"slices"
	"strings"
)

type Channel struct {
	Window
	topic *Topic
	users []*User
}

func NewChannel(connection *Connection, name string) *Channel {
	s, _ := uniqueid.Generateid("a", 5, "c")
	channel := &Channel{
		Window: Window{
			id:         s,
			name:       name,
			messages:   make([]*Message, 0),
			connection: connection,
		},
		topic: nil,
		users: nil,
	}
	return channel
}

func (c *Channel) SetTopic(topic *Topic) {
	c.topic = topic
}

func (c *Channel) GetTopic() *Topic {
	if c.topic == nil {
		return NewTopic("")
	}
	return c.topic
}

func (c *Channel) SetUsers(users []*User) {
	c.state.Lock()
	defer c.state.Unlock()
	c.users = users
	c.SortUsers()
}

func (c *Channel) AddUser(user *User) {
	c.state.Lock()
	defer c.state.Unlock()
	c.users = append(c.users, user)
	c.SortUsers()
}

func (c *Channel) GetUsers() []*User {
	var users []*User
	for i := range c.users {
		users = append(users, c.users[i])
	}
	return users
}

func (c *Channel) SortUsers() {
	slices.SortFunc(c.users, func(a, b *User) int {
		modeCmp := strings.Compare(b.modes, a.modes)
		if modeCmp != 0 {
			return modeCmp
		}
		return strings.Compare(a.nickname, b.nickname)
	})
}
