package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
)

func HandleUserModeSet(
	timestampFormat string,
	setPendingUpdate func(),
	setCurrentModes func(string),
	addMessage func(*Message),
) func(message ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		setCurrentModes(message.Params[1])
		addMessage(NewEvent(EventMode, timestampFormat, false, "Your modes changed: "+message.Params[1]))
	}
}
