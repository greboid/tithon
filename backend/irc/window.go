package irc

import (
	"slices"
	"strings"
	"sync"
)

type WindowState string

const (
	UnreadMessage   = "message"
	UnreadEvent     = "event"
	UnreadHighlight = "highlight"
	Read            = "read"
	Active          = "active"
)

type Window struct {
	id           string
	name         string
	title        string
	messages     []*Message
	connection   *Connection
	stateSync    sync.Mutex
	state        WindowState
	hasUsers     bool
	users        []*User
	isServer     bool
	tabCompleter TabCompleter
}

func (c *Window) GetID() string {
	return c.id
}

func (c *Window) GetName() string {
	c.stateSync.Lock()
	defer c.stateSync.Unlock()
	return c.name
}

func (c *Window) SetName(name string) {
	c.stateSync.Lock()
	defer c.stateSync.Unlock()
	c.name = name
}

func (c *Window) AddMessage(message *Message) {
	c.stateSync.Lock()
	defer c.stateSync.Unlock()
	c.messages = append(c.messages, message)
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

func (c *Window) GetState() string {
	c.stateSync.Lock()
	defer c.stateSync.Unlock()
	return string(c.state)
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
	if !c.hasUsers {
		return nil
	}
	var users []*User
	for i := range c.users {
		users = append(users, c.users[i])
	}
	return users
}

func (c *Window) SortUsers() {
	// TODO: Pull out info function
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

func (c *Window) IsServer() bool {
	return c.isServer
}

func (c *Window) GetTabCompleter() TabCompleter {
	return c.tabCompleter
}
