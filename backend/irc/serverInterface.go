package irc

import (
	"github.com/ergochat/irc-go/ircevent"
	"github.com/ergochat/irc-go/ircmsg"
)

type ServerInterface interface {
	GetID() string
	GetHostname() string
	Connect()
	Disconnect()
	IsServer() bool

	GetName() string
	SetName(name string)

	CurrentNick() string
	SetNick(nick string)

	HasCapability(name string) bool
	GetFileHost() string
	ISupport(value string) string
	SplitMessage(prefixLength int, message string) []string
	GetMaxLineLen() int

	AddConnectCallback(callback func(message ircmsg.Message))
	AddCallback(command string, callback func(ircmsg.Message))
	AddBatchCallback(callback func(batch *ircevent.Batch) bool)
	AddDisconnectCallback(callback func(message ircmsg.Message))

	IsValidChannel(target string) bool
	GetChannels() []*Channel
	GetChannel(id string) *Channel
	GetChannelByName(name string) (*Channel, error)
	AddChannel(name string) *Channel
	RemoveChannel(s string)
	JoinChannel(channel string, password string) error
	PartChannel(channel string) error
	IsChannel() bool

	GetQueries() []*Query
	GetQuery(id string) *Query
	GetQueryByName(name string) (*Query, error)
	AddQuery(name string) *Query
	RemoveQuery(id string)
	IsQuery() bool

	SendMessage(window string, message string) error
	SendQuery(target string, message string) error
	SendNotice(window string, message string) error
	SendQueryNotice(target string, message string) error
	SendRaw(message string)
	SendTopic(channel string, topic string) error

	GetCurrentModes() string
	SetCurrentModes(modes string)
	GetModePrefixes() []string
	GetModeNameForMode(mode string) string
	GetChannelModeType(mode string) rune

	GetCredentials() (string, string)
	AddMessage(message *Message)
	GetMessages() []*Message
	GetServer() ServerInterface
	SetActive(b bool)
	GetState() string
	GetTitle() string
	SetTitle(title string)
	GetTabCompleter() TabCompleter
	GetWindow() *Window
	SetWindowRemovalCallback(callback WindowRemovalCallback)
}
