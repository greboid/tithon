package irc

import (
	"fmt"
	"github.com/ergochat/irc-go/ircmsg"
	"regexp"
	"strings"
)

func HandleDisconnected(
	linkRegex *regexp.Regexp,
	timestampFormat string,
	setPendingUpdate setPendingUpdate,
	getQueries getQueries,
	getServerHostname getServerHostname,
	getChannels getChannels,
	addMessage addMessage,
) func(message ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		disconnectMessage := fmt.Sprintf("Disconnected from %s: %s", getServerHostname(), strings.Join(message.Params, " "))
		addMessage(NewEvent(linkRegex, EventDisconnected, timestampFormat, false, disconnectMessage))
		for _, channel := range getChannels() {
			channel.AddMessage(NewEvent(linkRegex, EventDisconnected, timestampFormat, false, disconnectMessage))
		}
		for _, query := range getQueries() {
			query.AddMessage(NewEvent(linkRegex, EventDisconnected, timestampFormat, false, disconnectMessage))
		}
	}
}
