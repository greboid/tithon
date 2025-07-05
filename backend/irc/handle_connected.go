package irc

import (
	"fmt"
	"github.com/ergochat/irc-go/ircmsg"
)

func HandleConnected(
	timestampFormat string,
	setPendingUpdate func(),
	getQueries func() []*Query,
	getServerHostname func() string,
	iSupport func(string) string,
	setServerName func(string),
	getChannels func() []*Channel,
	addMessage func(*Message),
) func(message ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		network := iSupport("NETWORK")
		if len(network) > 0 {
			setServerName(network)
		}
		connectMessage := fmt.Sprintf("Connected to %s", getServerHostname())
		addMessage(NewEvent(EventConnecting, timestampFormat, false, connectMessage))
		for _, channel := range getChannels() {
			channel.AddMessage(NewEvent(EventConnecting, timestampFormat, false, connectMessage))
		}
		for _, query := range getQueries() {
			query.AddMessage(NewEvent(EventConnecting, timestampFormat, false, connectMessage))
		}
	}
}
