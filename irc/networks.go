package irc

import (
	"fmt"
	"strings"
	"time"

	"github.com/ergochat/irc-go/ircevent"
	"github.com/ergochat/irc-go/ircmsg"
)

type Network struct {
	Name           string
	Profile        *Profile          `json:"profile"`
	Hostname       string            `json:"hostname"`
	UseTLS         bool              `json:"useTLS"`
	Channels       []*Channel        `json:"channels"`
	Queries        []*Query          `json:"queries"`
	StatusMessages []*NetworkMessage `json:"-"`
	connection     *ircevent.Connection
	updater        Updater
}

type Profile struct {
	Nickname     string `json:"nickname"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Ident        string `json:"ident"`
	RealName     string `json:"realname"`
	SASLUsername string `json:"saslusername"`
	SASLPassword string `json:"saslpassword"`
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
		Debug:       true,
	}
	n.addCallbacks()
	err := n.connection.Connect()
	if err != nil {
		fmt.Printf("Error connecting: %s", err.Error())
	}
}

func (n *Network) addCallbacks() {
	n.connection.AddConnectCallback(func(message ircmsg.Message) {
	})
	n.connection.AddCallback("JOIN", func(message ircmsg.Message) {
		if message.Nick() == n.connection.CurrentNick() {
			n.addToChannels(message.Params[0])
		}
	})
	n.connection.AddCallback("KICK", func(message ircmsg.Message) {
		if message.Params[1] == n.connection.CurrentNick() {
			n.removeFromChannels(message.Params[0])
		}
	})
	n.connection.AddCallback("PART", func(message ircmsg.Message) {
		if message.Nick() == n.connection.CurrentNick() {
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
			n.Channels[i].Joined = true
			existing = true
			break
		}
	}
	if !existing {
		newChannel := &Channel{Name: channel, Joined: true}
		n.Channels = append(n.Channels, newChannel)
	}
	n.updater.sendServerLists()
}

func (n *Network) removeFromChannels(channel string) {
	for i, v := range n.Channels {
		if v.Name == channel {
			n.Channels = append(n.Channels[:i], n.Channels[i+1:]...)
			break
		}
	}
	n.updater.sendServerLists()
}

func (n *Network) handlePrivMessage(message ircmsg.Message) {
	userHost, err := message.NUH()
	if err != nil {
		panic("Unable to parse NUH for host")
	}
	n.updater.sendChannelMessage(n, n.Channels[0], ChannelMessage{
		Source: User{
			Name: message.Nick(),
			Host: userHost.Host,
		},
		Time:    time.Now().Unix(),
		Message: strings.Join(message.Params[1:], " "),
	})
}
