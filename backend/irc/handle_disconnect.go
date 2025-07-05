package irc

import (
	"fmt"
	"github.com/ergochat/irc-go/ircmsg"
	"strings"
)

func HandleDisconnected(
	timestampFormat string,
	setPendingUpdate func(),
	getQueries func() []*Query,
	getServerHostname func() string,
	getChannels func() []*Channel,
	addMessage func(*Message),
) func(message ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		disconnectMessage := fmt.Sprintf("Disconnected from %s: %s", getServerHostname(), strings.Join(message.Params, " "))
		addMessage(NewEvent(EventDisconnected, timestampFormat, false, disconnectMessage))
		for _, channel := range getChannels() {
			channel.AddMessage(NewEvent(EventDisconnected, timestampFormat, false, disconnectMessage))
		}
		for _, query := range getQueries() {
			query.AddMessage(NewEvent(EventDisconnected, timestampFormat, false, disconnectMessage))
		}
	}
}
