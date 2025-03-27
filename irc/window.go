package irc

import (
	"slices"
	"strings"
	"sync"
	"sync/atomic"
)

type Window struct {
	id         string
	name       string
	title      string
	messages   []*Message
	connection *Connection
	state      sync.Mutex
	active     atomic.Bool
	unread     atomic.Bool
	hasUsers   atomic.Bool
	users      []*User
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

func (c *Window) SetUsers(users []*User) {
	c.state.Lock()
	defer c.state.Unlock()
	c.users = users
	c.SortUsers()
}

func (c *Window) AddUser(user *User) {
	c.state.Lock()
	defer c.state.Unlock()
	c.users = append(c.users, user)
	c.SortUsers()
}

func (c *Window) GetUsers() []*User {
	if !c.hasUsers.Load() {
		return nil
	}
	var users []*User
	for i := range c.users {
		users = append(users, c.users[i])
	}
	return users
}

func (c *Window) SortUsers() {
	slices.SortFunc(c.users, func(a, b *User) int {
		modeCmp := strings.Compare(b.modes, a.modes)
		if modeCmp != 0 {
			return modeCmp
		}
		return strings.Compare(a.nickname, b.nickname)
	})
}

func (c *Window) SetTitle(title string) {
	c.title = title
}

func (c *Window) GetTitle() string {
	return c.title
}
