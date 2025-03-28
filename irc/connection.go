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
	"time"
)

type Connection struct {
	*Window
	connection        *ircevent.Connection
	hostname          string
	port              int
	tls               bool
	password          string
	saslLogin         string
	saslPassword      string
	preferredNickname string
	channels          map[string]*Channel
	pms               map[string]*PrivateMessage
	mutex             sync.Mutex
	callbackHandler   *Handler
	supportsFileHost  bool
	currentModes      string
	possibleUserModes []*UserMode
}

func NewConnection(hostname string, port int, tls bool, password string, sasllogin string, saslpassword string, profile *Profile) *Connection {
	s, _ := uniqueid.Generateid("a", 5, "s")
	useSasl := len(sasllogin) > 0 && len(saslpassword) > 0

	connection := &Connection{
		hostname:          hostname,
		port:              port,
		tls:               tls,
		password:          password,
		saslLogin:         sasllogin,
		saslPassword:      saslpassword,
		preferredNickname: profile.nickname,
		channels:          map[string]*Channel{},
		pms:               map[string]*PrivateMessage{},
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
	}
	connection.Window = &Window{
		id:         s,
		name:       hostname,
		title:      hostname,
		messages:   make([]*Message, 0),
		connection: connection,
	}

	return connection
}

func (c *Connection) GetID() string {
	return c.id
}

func (c *Connection) GetFileHost() string {
	if c.connection == nil {
		return ""
	}
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
	c.AddMessage(NewEvent(time.Now(), fmt.Sprintf("Connecting to %s", c.connection.Server)))
	//TODO Need to store a connection state
	if !c.connection.Connected() {
		err := c.connection.Connect()
		if err != nil {
			c.AddMessage(NewError(time.Now(), "Connection error: "+err.Error()))
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

func (c *Connection) SendMessage(time time.Time, window string, message string) error {
	channel := c.GetChannel(window)
	if channel == nil {
		return errors.New("not on a channel")
	}
	if !c.HasCapability("echo-message") {
		channel.AddMessage(NewMessage(time, c.connection.CurrentNick(), message))
	}
	return c.connection.Send("PRIVMSG", channel.name, message)
}

func (c *Connection) SendNotice(time time.Time, window string, message string) error {
	channel := c.GetChannel(window)
	if channel == nil {
		return errors.New("not on a channel")
	}
	if !c.HasCapability("echo-message") {
		channel.AddMessage(NewMessage(time, c.connection.CurrentNick(), message))
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

func (c *Connection) GetModeNameForMode(mode string) string {
	modes := c.GetModePrefixes()
	index := strings.Index(modes[0], mode)
	if index == -1 {
		return ""
	}
	return modes[1][index : index+1]
}
