package irc

import uniqueid "github.com/albinj12/unique-id"

type Channel struct {
	id        string
	name      string
	messages  []*Message
	topic     *Topic
	users     []*User
	conection *Connection
	unread    bool
}

func NewChannel(connection *Connection, name string) *Channel {
	s, _ := uniqueid.Generateid("a", 5, "h")
	channel := &Channel{
		id:        s,
		conection: connection,
		name:      name,
		messages:  make([]*Message, 0),
	}
	return channel
}

func (c *Channel) GetID() string {
	return c.id
}

func (c *Channel) GetName() string {
	return c.name
}

func (c *Channel) AddMessage(message *Message) {
	c.messages = append(c.messages, message)
}

func (c *Channel) GetMessages() []*Message {
	var messages []*Message
	for _, message := range c.messages {
		messages = append(messages, message)
	}
	return messages
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

func (c *Channel) GetUsers() []string {
	var users []string
	for i := range c.users {
		users = append(users, c.users[i].nickname)
	}
	return users
}

func (c *Channel) GetServer() *Connection {
	return c.conection
}
