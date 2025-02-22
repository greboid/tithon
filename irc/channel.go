package irc

type Channel struct {
	id       string
	name     string
	messages []*Message
}

func (c *Channel) GetID() string {
	return c.id
}

func (c *Channel) GetName() string {
	return c.name
}

func (c *Channel) GetMessages() []string {
	var messages []string
	for _, message := range c.messages {
		messages = append(messages, message.GetMessage())
	}
	return messages
}
