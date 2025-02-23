package irc

import (
	"fmt"
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

func NewConnection(hostname string, port int, tls bool, sasllogin string, saslpassword string, profile *Profile) *Connection {
	s, _ := uniqueid.Generateid("a", 5, "s")
	useSasl := len(sasllogin) > 0 && len(saslpassword) > 0

	return &Connection{
		id:                s,
		hostname:          hostname,
		port:              port,
		tls:               tls,
		saslLogin:         sasllogin,
		saslPassword:      saslpassword,
		preferredNickname: profile.nickname,
		connection: &ircevent.Connection{
			Server:       fmt.Sprintf("%s:%d", hostname, port),
			Nick:         profile.nickname,
			SASLLogin:    sasllogin,
			SASLPassword: saslpassword,
			QuitMessage:  " ",
			Version:      " ",
			UseTLS:       tls,
			UseSASL:      useSasl,
			EnableCTCP:   false,
			Debug:        true,
		},
		channels: map[string]*Channel{},
	}
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
	c.connection.AddCallback("332", c.handleRPL_TOPIC)
	c.connection.AddCallback("TOPIC", c.handleTopic)
}

func (c *Connection) handleTopic(message ircmsg.Message) {
	slog.Debug("Handling topic", "message", message)
	for _, channel := range c.channels {
		if channel.name == message.Params[0] {
			topic := NewTopic(strings.Join(message.Params[1:], " "))
			slog.Debug("Setting topic", "server", c.GetName(), "channel", channel.GetName(), "topic", topic)
			channel.SetTopic(topic)
			return
		}
	}
}

func (c *Connection) handleRPL_TOPIC(message ircmsg.Message) {
	for _, channel := range c.channels {
		if channel.name == message.Params[1] {
			topic := NewTopic(strings.Join(message.Params[2:], " "))
			slog.Debug("Setting topic", "server", c.GetName(), "channel", channel.GetName(), "topic", topic)
			channel.SetTopic(topic)
			return
		}
	}
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
