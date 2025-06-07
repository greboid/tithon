package irc

import (
	"errors"
	"fmt"
	uniqueid "github.com/albinj12/unique-id"
	"github.com/ergochat/irc-go/ircevent"
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/greboid/tithon/config"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"sync"
	"time"
)

const (
	LevelTrace = slog.Level(-8)
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
	ut                UpdateTrigger
	nm                *NotificationManager
	conf              *config.Config
}

func NewConnection(conf *config.Config, id string, hostname string, port int, tls bool, password string, sasllogin string, saslpassword string, profile *Profile, ut UpdateTrigger, nm *NotificationManager) *Connection {
	if id == "" {
		id, _ = uniqueid.Generateid("a", 5, "s")
	}
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
			Timeout:       10 * time.Second,
			Server:        fmt.Sprintf("%s:%d", hostname, port),
			Nick:          profile.nickname,
			SASLLogin:     sasllogin,
			SASLPassword:  saslpassword,
			QuitMessage:   " ",
			Version:       " ",
			UseTLS:        tls,
			UseSASL:       useSasl,
			EnableCTCP:    false,
			ReconnectFreq: 5 * time.Second,
			RequestCaps: []string{
				"message-tags",
				"echo-message",
				"server-time",
				"soju.im/FILEHOST",
				"draft/chathistory",
				"draft/event-playback",
				"batch",
			},
			Debug: true,
			Log:   slog.NewLogLogger(slog.Default().Handler().WithAttrs([]slog.Attr{slog.Bool("rawirc", true), slog.String("Connection", id)}), LevelTrace),
		},
		ut:   ut,
		nm:   nm,
		conf: conf,
	}
	connection.Window = &Window{
		id:           id,
		name:         hostname,
		title:        hostname,
		messages:     make([]*Message, 0),
		connection:   connection,
		isServer:     true,
		tabCompleter: NewConnectionTabCompleter(connection),
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
	defer c.ut.SetPendingUpdate()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.callbackHandler == nil {
		c.callbackHandler = NewHandler(c)
		c.callbackHandler.addCallbacks()
	}
	c.AddMessage(NewEvent(c.conf.UISettings.TimestampFormat, false, fmt.Sprintf("Connecting to %s", c.connection.Server)))
	//TODO Need to store a connection state
	if !c.connection.Connected() {
		err := c.connection.Connect()
		if err != nil {
			c.AddMessage(NewError(c.conf.UISettings.TimestampFormat, false, "Connection error: "+err.Error()))
		}
	}
}

func (c *Connection) AddConnectCallback(callback func(message ircmsg.Message)) {
	c.connection.AddConnectCallback(callback)
}

func (c *Connection) AddCallback(command string, callback func(ircmsg.Message)) {
	c.connection.AddCallback(command, callback)
}

func (c *Connection) AddBatchCallback(callback func(batch *ircevent.Batch) bool) {
	c.connection.AddBatchCallback(callback)
}

func (c *Connection) GetCredentials() (string, string) {
	return c.saslLogin, c.saslPassword
}

func (c *Connection) Disconnect() {
	defer c.ut.SetPendingUpdate()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.connection.Quit()
}

func (c *Connection) IsChannel(target string) bool {
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

func (c *Connection) GetChannels() []*Channel {
	c.mutex.Lock()
	defer c.mutex.Unlock()
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
	defer c.ut.SetPendingUpdate()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	channel := NewChannel(c, name)
	c.channels[channel.id] = channel
	return channel
}

func (c *Connection) RemoveChannel(s string) {
	defer c.ut.SetPendingUpdate()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.PartChannel(s)
	delete(c.channels, s)
}

func (c *Connection) HasCapability(name string) bool {
	_, exists := c.connection.AcknowledgedCaps()[name]
	return exists
}

func (c *Connection) SendMessage(window string, message string) error {
	defer c.ut.SetPendingUpdate()
	channel := c.GetChannel(window)
	if channel == nil {
		return errors.New("not on a channel")
	}
	if !c.HasCapability("echo-message") {
		channel.AddMessage(NewMessage(c.conf.UISettings.TimestampFormat, true, c.connection.CurrentNick(), message, nil))
	}
	return c.connection.Send("PRIVMSG", channel.name, message)
}

func (c *Connection) SendNotice(window string, message string) error {
	channel := c.GetChannel(window)
	if channel == nil {
		return errors.New("not on a channel")
	}
	if !c.HasCapability("echo-message") {
		channel.AddMessage(NewMessage(c.conf.UISettings.TimestampFormat, true, c.connection.CurrentNick(), message, nil))
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

func (c *Connection) SendRaw(message string) {
	c.connection.SendRaw(message)
}

func (c *Connection) ISupport(value string) string {
	return c.connection.ISupport()[value]
}

func (c *Connection) GetHostname() string {
	return c.connection.Server
}

func (c *Connection) GetCurrentModes() string {
	return c.currentModes
}

func (c *Connection) SetCurrentModes(modes string) {
	c.currentModes = modes
}
