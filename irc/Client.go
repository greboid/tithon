package irc

import (
	"fmt"
	"github.com/ergochat/irc-go/ircevent"
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"newirc/gui"
)

const (
	v3TimestampFormat = "2006-01-02T15:04:05.000Z"
)

type Client struct {
	ircevent.Connection
	App *gui.App
}

func (client *Client) TestConnect() {

	client.SASLLogin = "login"
	client.SASLPassword = "password"
	client.Nick = "nickname"
	client.Server = "server"
	client.UseTLS = true
	client.AddConnectCallback(func(message ircmsg.Message) {
		fmt.Printf("Trying to join\n")
		err := client.Join("#MDBot")
		if err != nil {
			fmt.Printf("Join failed: %s\n", err.Error())
		}
	})
	client.AddCallback("PRIVMSG", func(message ircmsg.Message) {
		fmt.Println("Emitting event: %s", message.Nick())
		runtime.EventsEmit(client.App.Ctx, "privmsg", message)
	})
	client.Debug = true
	fmt.Printf("Trying to connect")
	err := client.Connect()
	if err != nil {
		fmt.Printf("Connect failed: %s\n", err.Error())
	}
	go func() {
		client.Loop()
	}()
}
