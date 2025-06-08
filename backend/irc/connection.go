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
	"strconv"
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
	reconnecting      bool
	reconnectAttempts int
	reconnectTimer    *time.Timer
	manualDisconnect  bool
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
				"batch",
			},
			Debug: true,
			Log:   slog.NewLogLogger(slog.Default().Handler().WithAttrs([]slog.Attr{slog.Bool("rawirc", true), slog.String("Connection", id)}), LevelTrace),
		},
		ut:                ut,
		nm:                nm,
		conf:              conf,
		reconnecting:      false,
		reconnectAttempts: 0,
		reconnectTimer:    nil,
		manualDisconnect:  false,
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
	c.manualDisconnect = false

	c.AddDisconnectCallback(func(message ircmsg.Message) {
		slog.Debug("Disconnected", "message", message)
		c.mutex.Lock()
		defer c.mutex.Unlock()
		if c.reconnecting || c.manualDisconnect {
			return
		}
		go c.scheduleReconnect()
	})
	c.AddConnectCallback(func(message ircmsg.Message) {
		c.mutex.Lock()
		defer c.mutex.Unlock()
		c.reconnecting = false
		c.reconnectAttempts = 0
		if c.reconnectTimer != nil {
			c.reconnectTimer.Stop()
			c.reconnectTimer = nil
		}
	})
	c.AddCallback("ERROR", func(message ircmsg.Message) {
		go c.scheduleReconnect()
	})

	c.AddMessage(NewEvent(c.conf.UISettings.TimestampFormat, false, fmt.Sprintf("Connecting to %s", c.connection.Server)))
	if !c.connection.Connected() {
		c.resetReconnectValues()
		err := c.connection.Connect()
		if err != nil {
			c.AddMessage(NewError(c.conf.UISettings.TimestampFormat, false, "Connection error: "+err.Error()))
			go c.scheduleReconnect()
		}
	}
}

func (c *Connection) resetReconnectValues() {
	c.reconnecting = false
	c.reconnectAttempts = 0
	if c.reconnectTimer != nil {
		c.reconnectTimer.Stop()
		c.reconnectTimer = nil
	}
}

func (c *Connection) scheduleReconnect() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	defer c.ut.SetPendingUpdate()
	if c.reconnecting {
		return
	}
	c.reconnecting = true
	c.reconnectAttempts++

	// TODO: Look at something better?
	delay := 1 << c.reconnectAttempts * time.Second
	if delay > 1*time.Minute {
		delay = 1 * time.Minute
	}

	c.AddMessage(NewEvent(c.conf.UISettings.TimestampFormat, false,
		fmt.Sprintf("Reconnection attempt %d scheduled in %v", c.reconnectAttempts, delay)))

	if c.reconnectTimer != nil {
		c.reconnectTimer.Stop()
	}

	c.reconnectTimer = time.AfterFunc(delay, func() {
		c.mutex.Lock()
		defer c.mutex.Unlock()
		defer c.ut.SetPendingUpdate()
		c.reconnecting = false

		c.AddMessage(NewEvent(c.conf.UISettings.TimestampFormat, false,
			fmt.Sprintf("Attempting to reconnect (attempt %d)...", c.reconnectAttempts)))

		if !c.connection.Connected() {
			err := c.connection.Connect()
			if err != nil {
				c.AddMessage(NewError(c.conf.UISettings.TimestampFormat, false,
					fmt.Sprintf("Reconnection attempt %d failed: %s", c.reconnectAttempts, err.Error())))
				go c.scheduleReconnect()
				return
			}
		}
	})
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

func (c *Connection) AddDisconnectCallback(callback func(message ircmsg.Message)) {
	c.connection.AddDisconnectCallback(callback)
}

func (c *Connection) GetCredentials() (string, string) {
	return c.saslLogin, c.saslPassword
}

func (c *Connection) Disconnect() {
	defer c.ut.SetPendingUpdate()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.manualDisconnect = true
	c.resetReconnectValues()
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

func (c *Connection) GetMaxLineLen() int {
	maxLineLen := 512 // Default IRC line length
	linelen := c.ISupport("LINELEN")
	if linelen != "" {
		if val, err := strconv.Atoi(linelen); err == nil && val > 0 {
			maxLineLen = val
		}
	}
	return maxLineLen
}

func (c *Connection) SplitMessage(prefixLength int, message string) []string {
	maxMsgLen := c.GetMaxLineLen() - prefixLength
	if len(message) <= maxMsgLen && !strings.Contains(message, "\n") {
		return []string{message}
	}
	var parts []string
	lines := strings.Split(message, "\n")

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		if len(line) <= maxMsgLen {
			parts = append(parts, line)
			continue
		}

		remainingLine := line
		for len(remainingLine) > 0 {
			if len(remainingLine) <= maxMsgLen {
				parts = append(parts, remainingLine)
				break
			}
			splitPos := maxMsgLen
			for splitPos > 0 && remainingLine[splitPos] != ' ' {
				splitPos--
			}
			if splitPos == 0 {
				splitPos = maxMsgLen
			}
			parts = append(parts, remainingLine[:splitPos])
			remainingLine = remainingLine[splitPos:]
			remainingLine = strings.TrimLeft(remainingLine, " ")
		}
	}

	return parts
}

func (c *Connection) SendMessage(window string, message string) error {
	defer c.ut.SetPendingUpdate()
	channel := c.GetChannel(window)
	if channel == nil {
		return errors.New("not on a channel")
	}

	//PRIVMSG #channel :message == 10 + channel name
	messageParts := c.SplitMessage(10+len(channel.name), message)

	for _, part := range messageParts {
		if !c.HasCapability("echo-message") {
			channel.AddMessage(NewMessage(c.conf.UISettings.TimestampFormat, true, c.connection.CurrentNick(), part, nil))
		}
		err := c.connection.Send("PRIVMSG", channel.name, part)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Connection) SendNotice(window string, message string) error {
	channel := c.GetChannel(window)
	if channel == nil {
		return errors.New("not on a channel")
	}

	//NOTICE #channel :message == 9 + channel name
	messageParts := c.SplitMessage(9+len(channel.name), message)

	for _, part := range messageParts {
		if !c.HasCapability("echo-message") {
			channel.AddMessage(NewMessage(c.conf.UISettings.TimestampFormat, true, c.connection.CurrentNick(), part, nil))
		}
		err := c.connection.Send("NOTICE", channel.name, part)
		if err != nil {
			return err
		}
	}

	return nil
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

// GetChannelModeType returns the type of channel mode
// A = modes that add/remove addresses
// B = modes that change settings with parameters
// C = modes that change settings only when set
// D = modes that change simple boolean settings
func (c *Connection) GetChannelModeType(mode string) rune {
	prefixes := c.GetModePrefixes()
	if strings.Contains(prefixes[0], mode) {
		return 'P'
	}

	//Get modes, use default
	chanModes := c.ISupport("CHANMODES")
	if chanModes == "" {
		chanModes = "beI,k,l,imnpst"
	}

	parts := strings.Split(chanModes, ",")
	if len(parts) != 4 {
		slog.Error("Invalid CHANMODES format", "CHANMODES", chanModes)
		return '?'
	}
	if strings.Contains(parts[0], mode) {
		return 'A'
	} else if strings.Contains(parts[1], mode) {
		return 'B'
	} else if strings.Contains(parts[2], mode) {
		return 'C'
	} else if strings.Contains(parts[3], mode) {
		return 'D'
	}
	return '?'
}
