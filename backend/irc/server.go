package irc

import (
	"errors"
	"fmt"
	uniqueid "github.com/albinj12/unique-id"
	"github.com/ergochat/irc-go/ircevent"
	"github.com/ergochat/irc-go/ircmsg"
	"log/slog"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	LevelTrace = slog.Level(-8)
)

type Server struct {
	*Window
	connection            *ircevent.Connection
	hostname              string
	port                  int
	tls                   bool
	password              string
	saslLogin             string
	saslPassword          string
	preferredNickname     string
	channels              map[string]*Channel
	pms                   map[string]*Query
	mutex                 sync.Mutex
	currentModes          string
	ut                    UpdateTrigger
	nm                    NotificationManager
	timestampFormat       string
	reconnecting          bool
	reconnectAttempts     int
	reconnectTimer        *time.Timer
	manualDisconnect      bool
	linkRegex             *regexp.Regexp
	windowRemovalCallback WindowRemovalCallback
}

func (c *Server) GetWindow() *Window {
	return c.Window
}

func NewServer(timestampFormat string, id string, hostname string, port int, tls bool, password string, sasllogin string, saslpassword string, profile *Profile, ut UpdateTrigger, nm NotificationManager) *Server {
	if id == "" {
		id, _ = uniqueid.Generateid("a", 5, "s")
	}
	useSasl := len(sasllogin) > 0 && len(saslpassword) > 0

	server := &Server{
		hostname:          hostname,
		port:              port,
		tls:               tls,
		password:          password,
		saslLogin:         sasllogin,
		saslPassword:      saslpassword,
		preferredNickname: profile.nickname,
		channels:          map[string]*Channel{},
		pms:               map[string]*Query{},
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
			Log:   slog.NewLogLogger(slog.Default().Handler().WithAttrs([]slog.Attr{slog.Bool("rawirc", true), slog.String("Server", id)}), LevelTrace),
		},
		ut:                ut,
		nm:                nm,
		timestampFormat:   timestampFormat,
		reconnecting:      false,
		reconnectAttempts: 0,
		reconnectTimer:    nil,
		manualDisconnect:  false,
		linkRegex:         linkRegex,
	}
	server.Window = &Window{
		id:           id,
		name:         hostname,
		title:        hostname,
		messages:     make([]*Message, 0),
		connection:   server,
		isServer:     true,
		tabCompleter: NewServerTabCompleter(server),
	}

	return server
}

func (c *Server) GetID() string {
	return c.id
}

func (c *Server) GetFileHost() string {
	if c.connection == nil {
		return ""
	}
	return c.connection.ISupport()["soju.im/FILEHOST"]
}

func (c *Server) Connect() {
	defer c.ut.SetPendingUpdate()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	AddCallbacks(c, c.ut, c.nm, c.timestampFormat)
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

	c.AddMessage(NewEvent(EventConnecting, c.timestampFormat, false, fmt.Sprintf("Connecting to %s", c.connection.Server)))
	if !c.connection.Connected() {
		c.resetReconnectValues()
		err := c.connection.Connect()
		if err != nil {
			c.AddMessage(NewError(c.timestampFormat, false, "Server error: "+err.Error()))
			go c.scheduleReconnect()
		}
	}
}

func (c *Server) resetReconnectValues() {
	c.reconnecting = false
	c.reconnectAttempts = 0
	if c.reconnectTimer != nil {
		c.reconnectTimer.Stop()
		c.reconnectTimer = nil
	}
}

func (c *Server) CancelReconnection() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	defer c.ut.SetPendingUpdate()
	c.resetReconnectValues()
}

func (c *Server) scheduleReconnect() {
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

	c.AddMessage(NewEvent(EventConnecting, c.timestampFormat, false, fmt.Sprintf("Reconnection attempt %d scheduled in %v", c.reconnectAttempts, delay)))

	if c.reconnectTimer != nil {
		c.reconnectTimer.Stop()
	}

	c.reconnectTimer = time.AfterFunc(delay, func() {
		c.mutex.Lock()
		defer c.mutex.Unlock()
		defer c.ut.SetPendingUpdate()
		c.reconnecting = false

		c.AddMessage(NewEvent(EventConnecting, c.timestampFormat, false, fmt.Sprintf("Attempting to reconnect (attempt %d)...", c.reconnectAttempts)))

		if !c.connection.Connected() {
			err := c.connection.Connect()
			if err != nil {
				c.AddMessage(NewError(c.timestampFormat, false,
					fmt.Sprintf("Reconnection attempt %d failed: %s", c.reconnectAttempts, err.Error())))
				go c.scheduleReconnect()
				return
			}
		}
	})
}

func (c *Server) AddConnectCallback(callback func(message ircmsg.Message)) {
	c.connection.AddConnectCallback(callback)
}

func (c *Server) AddCallback(command string, callback func(ircmsg.Message)) {
	c.connection.AddCallback(command, callback)
}

func (c *Server) AddBatchCallback(callback func(batch *ircevent.Batch) bool) {
	c.connection.AddBatchCallback(callback)
}

func (c *Server) AddDisconnectCallback(callback func(message ircmsg.Message)) {
	c.connection.AddDisconnectCallback(callback)
}

func (c *Server) GetCredentials() (string, string) {
	return c.saslLogin, c.saslPassword
}

func (c *Server) Disconnect() {
	defer c.ut.SetPendingUpdate()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.manualDisconnect = true
	c.resetReconnectValues()
	c.connection.Quit()
}

func (c *Server) IsTargetChannel(target string) bool {
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

func (c *Server) IsValidChannel(target string) bool {
	return c.IsTargetChannel(target)
}

func (c *Server) IsChannel() bool {
	return c.Window.IsChannel()
}

func (c *Server) GetChannels() []*Channel {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	channels := slices.Collect(maps.Values(c.channels))
	slices.SortStableFunc(channels, func(a, b *Channel) int {
		return strings.Compare(strings.ToLower(a.name), strings.ToLower(b.name))
	})
	return channels
}

func (c *Server) GetChannel(id string) *Channel {
	return c.channels[id]
}

func (c *Server) GetChannelByName(name string) (*Channel, error) {
	for _, channel := range c.GetChannels() {
		if strings.EqualFold(channel.name, name) {
			return channel, nil
		}
	}
	return nil, errors.New("channel not found")
}

func (c *Server) AddChannel(name string) *Channel {
	defer c.ut.SetPendingUpdate()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	channel := NewChannel(c, name)
	c.channels[channel.id] = channel
	return channel
}

func (c *Server) RemoveChannel(s string) {
	defer c.ut.SetPendingUpdate()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	channel := c.channels[s]
	if channel != nil && c.windowRemovalCallback != nil {
		c.windowRemovalCallback.OnWindowRemoved(channel.Window)
	}
	c.PartChannel(s)
	delete(c.channels, s)
}

func (c *Server) GetQueries() []*Query {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	pms := slices.Collect(maps.Values(c.pms))
	slices.SortStableFunc(pms, func(a, b *Query) int {
		return strings.Compare(strings.ToLower(a.name), strings.ToLower(b.name))
	})
	return pms
}

func (c *Server) GetQuery(id string) *Query {
	return c.pms[id]
}

func (c *Server) GetQueryByName(name string) (*Query, error) {
	for _, pm := range c.GetQueries() {
		if strings.EqualFold(pm.name, name) {
			return pm, nil
		}
	}
	return nil, errors.New("query not found")
}

func (c *Server) AddQuery(name string) *Query {
	defer c.ut.SetPendingUpdate()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	pm := NewQuery(c, name)
	c.pms[pm.id] = pm
	return pm
}

func (c *Server) RemoveQuery(id string) {
	defer c.ut.SetPendingUpdate()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	query := c.pms[id]
	if query != nil && c.windowRemovalCallback != nil {
		c.windowRemovalCallback.OnWindowRemoved(query.Window)
	}
	delete(c.pms, id)
}

func (c *Server) HasCapability(name string) bool {
	_, exists := c.connection.AcknowledgedCaps()[name]
	return exists
}

func (c *Server) GetMaxLineLen() int {
	maxLineLen := 512 // Default IRC line length
	linelen := c.ISupport("LINELEN")
	if linelen != "" {
		if val, err := strconv.Atoi(linelen); err == nil && val > 0 {
			maxLineLen = val
		}
	}
	return maxLineLen
}

func (c *Server) SplitMessage(prefixLength int, message string) []string {
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

func (c *Server) SendMessage(window string, message string) error {
	defer c.ut.SetPendingUpdate()
	channel := c.GetChannel(window)
	if channel == nil {
		pm := c.GetQuery(window)
		if pm == nil {
			return errors.New("not on a channel or in a query")
		}
		return c.SendQuery(pm.name, message)
	}

	// PRIVMSG #channel :message == 10 + channel name
	messageParts := c.SplitMessage(10+len(channel.name), message)

	for _, part := range messageParts {
		if !c.HasCapability("echo-message") {
			channel.AddMessage(NewMessage(c.timestampFormat, true, c.connection.CurrentNick(), part, nil))
		}
		err := c.connection.Send("PRIVMSG", channel.name, part)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Server) SendQuery(target string, message string) error {
	defer c.ut.SetPendingUpdate()
	pm, err := c.GetQueryByName(target)
	if err != nil {
		pm = c.AddQuery(target)
	}

	// PRIVMSG nickname :message == 10 + nickname
	messageParts := c.SplitMessage(10+len(target), message)

	for _, part := range messageParts {
		if !c.HasCapability("echo-message") {
			pm.AddMessage(NewMessage(c.timestampFormat, true, c.connection.CurrentNick(), part, nil))
		}
		err = c.connection.Send("PRIVMSG", target, part)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Server) SendNotice(window string, message string) error {
	channel := c.GetChannel(window)
	if channel == nil {
		pm := c.GetQuery(window)
		if pm == nil {
			return errors.New("not on a channel or in a query")
		}
		return c.SendQueryNotice(pm.name, message)
	}

	// NOTICE #channel :message == 9 + channel name
	messageParts := c.SplitMessage(9+len(channel.name), message)

	for _, part := range messageParts {
		if !c.HasCapability("echo-message") {
			channel.AddMessage(NewNotice(c.timestampFormat, true, c.connection.CurrentNick(), part, nil))
		}
		err := c.connection.Send("NOTICE", channel.name, part)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Server) SendQueryNotice(target string, message string) error {
	defer c.ut.SetPendingUpdate()
	pm, err := c.GetQueryByName(target)
	if err != nil {
		pm = c.AddQuery(target)
	}

	// NOTICE nickname :message == 9 + nickname
	messageParts := c.SplitMessage(9+len(target), message)

	for _, part := range messageParts {
		if !c.HasCapability("echo-message") {
			pm.AddMessage(NewNotice(c.timestampFormat, true, c.connection.CurrentNick(), part, nil))
		}
		err = c.connection.Send("NOTICE", target, part)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Server) CurrentNick() string {
	return c.connection.CurrentNick()
}

func (c *Server) JoinChannel(channel string, password string) error {
	return c.connection.Join(channel)
}

func (c *Server) PartChannel(channel string) error {
	channelInstance := c.GetChannel(channel)
	if channelInstance != nil {
		return c.connection.Part(channelInstance.GetName())
	}
	return fmt.Errorf("channel %s not found", channel)
}

func (c *Server) GetModePrefixes() []string {
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

func (c *Server) GetModeNameForMode(mode string) string {
	modes := c.GetModePrefixes()
	index := strings.Index(modes[0], mode)
	if index == -1 {
		return ""
	}
	return modes[1][index : index+1]
}

func (c *Server) SendRaw(message string) {
	c.connection.SendRaw(message)
}

func (c *Server) ISupport(value string) string {
	return c.connection.ISupport()[value]
}

func (c *Server) GetHostname() string {
	return c.connection.Server
}

func (c *Server) SetNick(nick string) {
	c.connection.SetNick(nick)
}

func (c *Server) SendTopic(channel string, topic string) error {
	return c.connection.Send("TOPIC", channel, topic)
}

func (c *Server) GetCurrentModes() string {
	return c.currentModes
}

func (c *Server) SetCurrentModes(modes string) {
	c.currentModes = modes
}

// GetChannelModeType returns the type of channel mode
// A = modes that add/remove addresses
// B = modes that change settings with parameters
// C = modes that change settings only when set
// D = modes that change simple boolean settings
func (c *Server) GetChannelModeType(mode string) rune {
	prefixes := c.GetModePrefixes()
	if strings.Contains(prefixes[0], mode) {
		return 'P'
	}

	// Get modes, use default
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

func (c *Server) SetWindowRemovalCallback(callback WindowRemovalCallback) {
	c.windowRemovalCallback = callback
}
