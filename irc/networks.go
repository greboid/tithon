package irc

import (
	"strings"
	"time"

	"github.com/ergochat/irc-go/ircevent"
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/ergochat/irc-go/ircutils"
)

type Network struct {
	Name           string
	Profile        *Profile          `yaml:"profile"`
	Hostname       string            `yaml:"hostname"`
	UseTLS         bool              `yaml:"useTLS"`
	Channels       []*Channel        `yaml:"channels"`
	Queries        []*Query          `yaml:"queries"`
	StatusMessages []*NetworkMessage `yaml:"-"`
	connection     *ircevent.Connection
	updater        Updater
}

type Profile struct {
	Nickname     string `yaml:"nickname"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	Ident        string `yaml:"ident"`
	RealName     string `yaml:"realname"`
	SASLUsername string `yaml:"saslusername"`
	SASLPassword string `yaml:"saslpassword"`
}

type NetworkMessage struct {
	Source  string
	Time    time.Time
	Message string
}

type User struct {
	Name string
	Host string
}

type Updater interface {
	sendServerLists()
	sendChannelMessage(network *Network, channel *Channel, message ChannelMessage)
}

func (n *Network) Connect(updater Updater) {
	n.updater = updater
	n.connection = &ircevent.Connection{
		Server:      n.Hostname,
		Nick:        n.Profile.Nickname,
		User:        n.Profile.Username,
		RealName:    n.Profile.RealName,
		Password:    n.Profile.Password,
		QuitMessage: " ",
		Version:     "",
		UseTLS:      n.UseTLS,
		Debug:       false,
	}
	n.addCallbacks()
	_ = n.connection.Connect()
}

func (n *Network) addCallbacks() {
	n.connection.AddConnectCallback(func(message ircmsg.Message) {
	})
	n.connection.AddCallback("JOIN", func(message ircmsg.Message) {
		if ircutils.ParseUserhost(message.Prefix).Nick == n.connection.CurrentNick() {
			n.addToChannels(message.Params[0])
		}
	})
	n.connection.AddCallback("KICK", func(message ircmsg.Message) {
		if message.Params[1] == n.connection.CurrentNick() {
			n.removeFromChannels(message.Params[0])
		}
	})
	n.connection.AddCallback("PART", func(message ircmsg.Message) {
		if ircutils.ParseUserhost(message.Prefix).Nick == n.connection.CurrentNick() {
			n.removeFromChannels(message.Params[0])
		}
	})
	n.connection.AddCallback("PRIVMSG", func(message ircmsg.Message) {
		n.handlePrivMessage(message)
	})
}

func (n *Network) addToChannels(channel string) {
	existing := false
	for i := range n.Channels {
		if n.Channels[i].Name == channel {
			existing = true
			break
		}
	}
	if !existing {
		n.Channels = append(n.Channels, &Channel{Name: channel})
		n.updater.sendServerLists()
	}
}

func (n *Network) removeFromChannels(channel string) {
	for i, v := range n.Channels {
		if v.Name == channel {
			n.Channels = append(n.Channels[:i], n.Channels[i+1:]...)
			n.updater.sendServerLists()
			break
		}
	}
}

func (n *Network) handlePrivMessage(message ircmsg.Message) {
	userHost := ircutils.ParseUserhost(message.Prefix)
	n.updater.sendChannelMessage(n, n.Channels[0], ChannelMessage{
		Source: User{
			Name: userHost.Nick,
			Host: userHost.Host,
		},
		Time:    time.Now().Unix(),
		Message: strings.Join(message.Params[1:], " "),
	})
}
