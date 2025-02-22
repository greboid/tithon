package irc

import (
	uniqueid "github.com/albinj12/unique-id"
	"github.com/ergochat/irc-go/ircevent"
	"github.com/ergochat/irc-go/ircmsg"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"sync"
)

type Connection struct {
	id                string
	hostname          string
	port              int
	tls               bool
	saslLogin         string
	saslPassword      string
	preferredNickname string
	channels          map[string]*Channel
	connection        *ircevent.Connection
	callbacksAdded    bool
	mutex             sync.Mutex
}

func (c *Connection) GetID() string {
	return c.id
}

func (c *Connection) GetName() string {
	network := c.connection.ISupport()["NETWORK"]
	if c.connection.Connected() && len(network) > 0 {
		return network
	}
	return c.hostname
}

func (c *Connection) Connect() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if !c.callbacksAdded {
		c.addCallbacks()
		c.callbacksAdded = true
	}
	if !c.connection.Connected() {
		c.connection.Connect()
	}

}

func (c *Connection) Disconnect() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.connection.Quit()
}

func (c *Connection) GetChannels() []*Channel {
	channels := slices.Collect(maps.Values(c.channels))
	slices.SortStableFunc(channels, func(a, b *Channel) int {
		return strings.Compare(strings.ToLower(a.name), strings.ToLower(b.name))
	})
	return channels
}

func (c *Connection) GetChannel(id string) *Channel {
	return c.channels[id]
}

func (c *Connection) isChannel(target string) bool {
	chanTypes := c.connection.ISupport()["CHANTYPES"]
	if chanTypes == "" {
		chanTypes = "#"
	}
	for _, char := range chanTypes {
		if strings.HasPrefix(target, string(char)) {
			return true
		}
	}
	return false
}

func (c *Connection) SendMessage(window string, message string) {
	channel := c.GetChannel(window)
	if channel == nil {
		return
	}
	if c.connection.AcknowledgedCaps()["echo-message"] == "" {
		channel.messages = append(channel.messages, NewMessage(c.connection.CurrentNick(), message))
	}
	err := c.connection.Send("PRIVMSG", channel.name, message)
	if err != nil {
		slog.Error("Unable to send message", "server", c.GetName(), "channel", channel.name, "message", message)
	}
}

func (c *Connection) addCallbacks() {
	c.connection.AddCallback("JOIN", c.handleJoin)
	c.connection.AddCallback("PRIVMSG", c.handlePrivMsg)
}

func (c *Connection) handlePrivMsg(message ircmsg.Message) {
	mess := NewMessage(message.Nick(), strings.Join(message.Params[1:], " "))
	if c.isChannel(message.Params[0]) {
		for _, channel := range c.channels {
			if channel.name == message.Params[0] {
				channel.messages = append(channel.messages, mess)
				return
			}
		}
		slog.Warn("Message for unknown channel", "message", message)
	} else {
		slog.Warn("Unsupported DM", "message", message)
	}
}

func (c *Connection) handleJoin(message ircmsg.Message) {
	if message.Nick() == c.connection.CurrentNick() {
		c.handleSelfJoin(message)
	} else {
		c.handleOtherJoin(message)
	}
}

func (c *Connection) handleSelfJoin(message ircmsg.Message) {
	slog.Debug("Joining channel", "channel", message.Params[0])
	s, _ := uniqueid.Generateid("a", 5, "c")
	c.channels[s] = &Channel{
		id:   s,
		name: message.Params[0],
	}
}

func (c *Connection) handleOtherJoin(message ircmsg.Message) {

}
