package irc

import (
	"fmt"
	"github.com/ergochat/irc-go/ircmsg"
	"regexp"
)

func HandleConnected(
	linkRegex *regexp.Regexp,
	timestampFormat string,
	setPendingUpdate setPendingUpdate,
	getQueries getQueries,
	getServerHostname getServerHostname,
	iSupport iSupport,
	setServerName setServerName,
	getChannels getChannels,
	addMessage addMessage,
) func(message ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		network := iSupport("NETWORK")
		if len(network) > 0 {
			setServerName(network)
		}
		connectMessage := fmt.Sprintf("Connected to %s", getServerHostname())
		addMessage(NewEvent(linkRegex, EventConnecting, timestampFormat, false, connectMessage))
		for _, channel := range getChannels() {
			channel.AddMessage(NewEvent(linkRegex, EventConnecting, timestampFormat, false, connectMessage))
		}
		for _, query := range getQueries() {
			query.AddMessage(NewEvent(linkRegex, EventConnecting, timestampFormat, false, connectMessage))
		}
	}
}
