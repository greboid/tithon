package irc

import (
	"slices"
	"strings"
	"sync"
	"sync/atomic"
)

type WindowState string

const (
	UnreadMessage   = "unread"
	UnreadEvent     = "unread event"
	UnreadHighlight = "unread highlight"
	Read            = "read"
	Active          = "active"
)

type Window struct {
	id         string
	name       string
	title      string
	messages   []*Message
	connection *Connection
	stateSync  sync.Mutex
	state      WindowState
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
	c.stateSync.Lock()
	defer c.stateSync.Unlock()
	if c.state == Active {
		return
	}
	switch message.messageType {
	case Error, Event:
		c.state = UnreadEvent
	case Normal, Notice, Action:
		c.state = UnreadMessage
	case Highlight, HighlightNotice, HighlightAction:
		c.state = UnreadHighlight
	}
	c.messages = append(c.messages, message)
}

func (c *Window) GetMessages() []*Message {
	c.stateSync.Lock()
	defer c.stateSync.Unlock()
	var messages []*Message
	for _, message := range c.messages {
		messages = append(messages, message)
	}
	return messages
}

func (c *Window) GetServer() *Connection {
	return c.connection
}

func (c *Window) SetActive(b bool) {
	c.stateSync.Lock()
	defer c.stateSync.Unlock()
	if b {
		c.state = Active
	} else {
		c.state = Read
	}
}

func (c *Window) IsActive() bool {
	c.stateSync.Lock()
	defer c.stateSync.Unlock()
	return c.state == Active
}

func (c *Window) IsUnread() bool {
	c.stateSync.Lock()
	defer c.stateSync.Unlock()
	return c.state == UnreadMessage || c.state == UnreadEvent || c.state == UnreadHighlight
}

func (c *Window) SetUsers(users []*User) {
	c.stateSync.Lock()
	defer c.stateSync.Unlock()
	c.users = users
	c.SortUsers()
}

func (c *Window) AddUser(user *User) {
	c.stateSync.Lock()
	defer c.stateSync.Unlock()
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
