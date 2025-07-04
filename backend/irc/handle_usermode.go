package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"regexp"
)

func HandleUserModeSet(
	linkRegex *regexp.Regexp,
	timestampFormat string,
	setPendingUpdate setPendingUpdate,
	setCurrentModes setCurrentModes,
	addMessage addMessage,
) func(message ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		setCurrentModes(message.Params[1])
		addMessage(NewEvent(linkRegex, EventMode, timestampFormat, false, "Your modes changed: "+message.Params[1]))
	}
}
