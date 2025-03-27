package irc

import (
	uniqueid "github.com/albinj12/unique-id"
	"sync"
	"sync/atomic"
)

type PrivateMessage struct {
	id        string
	user      *User
	messages  []*Message
	conection *Connection
	state     sync.Mutex
	active    atomic.Bool
	unread    atomic.Bool
}

func NewPrivateMessage(connection *Connection, name string) *PrivateMessage {
	s, _ := uniqueid.Generateid("a", 5, "p")
	channel := &PrivateMessage{
		id:        s,
		conection: connection,
		user:      NewUser(name, ""),
		messages:  make([]*Message, 0),
	}
	return channel
}

func (c *PrivateMessage) GetName() string {
	return c.user.nickname
}

func (c *PrivateMessage) AddMessage(message *Message) {
	if !c.active.Load() {
		c.unread.Store(true)
	}
	c.state.Lock()
	c.messages = append(c.messages, message)
	c.state.Unlock()
}

func (c *PrivateMessage) GetMessages() []*Message {
	var messages []*Message
	c.state.Lock()
	for _, message := range c.messages {
		messages = append(messages, message)
	}
	c.state.Unlock()
	return messages
}

func (c *PrivateMessage) GetServer() *Connection {
	return c.conection
}

func (c *PrivateMessage) SetActive(b bool) {
	c.active.Store(b)
}

func (c *PrivateMessage) IsActive() bool {
	return c.active.Load()
}

func (c *PrivateMessage) SetUnread(b bool) {
	c.unread.Store(b)
}

func (c *PrivateMessage) IsUnread() bool {
	return c.unread.Load()
}
