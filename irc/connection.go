package irc

import (
	"errors"
	"fmt"
	uniqueid "github.com/albinj12/unique-id"
	"github.com/ergochat/irc-go/ircevent"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Connection struct {
	id                string
	hostname          string
	port              int
	tls               bool
	password          string
	saslLogin         string
	saslPassword      string
	preferredNickname string
	channels          map[string]*Channel
	connection        *ircevent.Connection
	mutex             sync.Mutex
	callbackHandler   *Handler
	supportsFileHost  bool
	currentModes      string
	messageLock       sync.Mutex
	messages          []*Message
	unread            atomic.Bool
	active            atomic.Bool
}

func NewConnection(hostname string, port int, tls bool, password string, sasllogin string, saslpassword string, profile *Profile) *Connection {
	s, _ := uniqueid.Generateid("a", 5, "s")
	useSasl := len(sasllogin) > 0 && len(saslpassword) > 0

	return &Connection{
		id:                s,
		hostname:          hostname,
		port:              port,
		tls:               tls,
		password:          password,
		saslLogin:         sasllogin,
		saslPassword:      saslpassword,
		preferredNickname: profile.nickname,
		connection: &ircevent.Connection{
			Timeout:      10 * time.Second,
			Server:       fmt.Sprintf("%s:%d", hostname, port),
			Nick:         profile.nickname,
			SASLLogin:    sasllogin,
			SASLPassword: saslpassword,
			QuitMessage:  " ",
			Version:      " ",
			UseTLS:       tls,
			UseSASL:      useSasl,
			EnableCTCP:   false,
			RequestCaps: []string{
				"message-tags",
				"echo-message",
				"server-time",
				"soju.im/FILEHOST",
				"draft/chathistory",
				"draft/event-playback",
			},
			Debug: true,
		},
		channels: map[string]*Channel{},
		messages: make([]*Message, 0),
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

func (c *Connection) GetFileHost() string {
	return c.connection.ISupport()["soju.im/FILEHOST"]
}

func (c *Connection) Connect() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.callbackHandler == nil {
		c.callbackHandler = &Handler{
			connection: c,
		}
		c.callbackHandler.addCallbacks()
	}
	c.AddMessage(NewMessage("", fmt.Sprintf("Connecting to %s", c.connection.Server), Event))
	//TODO Need to store a connection state
	if !c.connection.Connected() {
		err := c.connection.Connect()
		if err != nil {
			c.AddMessage(NewMessage("", "Connection error: "+err.Error(), Event))
		}
	}

}

func (c *Connection) GetCredentials() (string, string) {
	return c.saslLogin, c.saslPassword
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

func (c *Connection) GetChannelByName(name string) (*Channel, error) {
	for _, channel := range c.GetChannels() {
		if strings.ToLower(channel.name) == strings.ToLower(name) {
			return channel, nil
		}
	}
	return nil, errors.New("channel not found")
}

func (c *Connection) AddChannel(name string) *Channel {
	channel := NewChannel(c, name)
	c.channels[channel.id] = channel
	return channel
}

func (c *Connection) RemoveChannel(s string) {
	c.PartChannel(s)
	delete(c.channels, s)
}

func (c *Connection) HasCapability(name string) bool {
	_, exists := c.connection.AcknowledgedCaps()[name]
	return exists
}

func (c *Connection) SendMessage(window string, message string) error {
	channel := c.GetChannel(window)
	if channel == nil {
		return errors.New("not on a channel")
	}
	if !c.HasCapability("echo-message") {
		channel.AddMessage(NewMessage(c.connection.CurrentNick(), message, Normal))
	}
	return c.connection.Send("PRIVMSG", channel.name, message)
}

func (c *Connection) SendNotice(window string, message string) error {
	channel := c.GetChannel(window)
	if channel == nil {
		return errors.New("not on a channel")
	}
	if !c.HasCapability("echo-message") {
		channel.AddMessage(NewMessage(c.connection.CurrentNick(), message, Notice))
	}
	return c.connection.Send("NOTICE", channel.name, message)
}

func (c *Connection) CurrentNick() string {
	return c.connection.CurrentNick()
}

func (c *Connection) JoinChannel(channel string, password string) error {
	return c.connection.Join(channel)
}

func (c *Connection) PartChannel(channel string) error {
	return c.connection.Part(c.GetChannel(channel).GetName())
}

func (c *Connection) GetModePrefixes() []string {
	value, exists := c.connection.ISupport()["PREFIX"]
	if !exists {
		slog.Error("No mode prefixes specified, using default")
		value = "(o)@"
	}
	splits := strings.Split(value[1:], ")")
	if len(splits[0]) != len(splits[1]) {
		slog.Error("Error parsing mode prefixes", "PREFIX", value)
		splits[0] = "o"
		splits[1] = "@"
	}
	return splits
}

func (c *Connection) AddMessage(message *Message) {
	if !c.active.Load() {
		c.unread.Store(true)
	}
	c.messageLock.Lock()
	c.messages = append(c.messages, message)
	c.messageLock.Unlock()
}

func (c *Connection) GetMessages() []*Message {
	var messages []*Message
	c.messageLock.Lock()
	for _, message := range c.messages {
		messages = append(messages, message)
	}
	c.messageLock.Unlock()
	return messages
}

func (c *Connection) SetActive(b bool) {
	c.active.Store(b)
}

func (c *Connection) IsActive() bool {
	return c.active.Load()
}

func (c *Connection) SetUnread(b bool) {
	c.unread.Store(b)
}

func (c *Connection) IsUnread() bool {
	return c.unread.Load()
}
