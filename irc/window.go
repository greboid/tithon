package irc

import (
	"sync"
	"sync/atomic"
)

type Window struct {
	id         string
	name       string
	messages   []*Message
	connection *Connection
	state      sync.Mutex
	active     atomic.Bool
	unread     atomic.Bool
}

func (c *Window) GetID() string {
	return c.id
}

func (c *Window) GetName() string {
	return c.name
}

func (c *Window) SetName(name string) {
	c.name = name
}

func (c *Window) AddMessage(message *Message) {
	if !c.active.Load() {
		c.unread.Store(true)
	}
	c.state.Lock()
	c.messages = append(c.messages, message)
	c.state.Unlock()
}

func (c *Window) GetMessages() []*Message {
	var messages []*Message
	c.state.Lock()
	for _, message := range c.messages {
		messages = append(messages, message)
	}
	c.state.Unlock()
	return messages
}

func (c *Window) GetServer() *Connection {
	return c.connection
}

func (c *Window) SetActive(b bool) {
	c.active.Store(b)
}

func (c *Window) IsActive() bool {
	return c.active.Load()
}

func (c *Window) SetUnread(b bool) {
	c.unread.Store(b)
}

func (c *Window) IsUnread() bool {
	return c.unread.Load()
}
