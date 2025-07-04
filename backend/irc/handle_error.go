package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"regexp"
	"strings"
)

func HandleError(
	linkRegex *regexp.Regexp,
	timestampFormat string,
	setPendingUpdate setPendingUpdate,
	addMessage addMessage,
) func(ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		addMessage(NewError(linkRegex, timestampFormat, false, strings.Join(message.Params, " ")))
	}
}

func HandleNickInUse(
	linkRegex *regexp.Regexp,
	timestampFormat string,
	setPendingUpdate setPendingUpdate,
	addMessage addMessage,
) func(ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		addMessage(NewError(linkRegex, timestampFormat, false, "Nickname in use: "+message.Params[1]))
	}
}

func HandlePasswordMismatch(
	linkRegex *regexp.Regexp,
	timestampFormat string,
	setPendingUpdate setPendingUpdate,
	addMessage addMessage,
) func(ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		addMessage(NewError(linkRegex, timestampFormat, false, "Password Mismatch: "+strings.Join(message.Params, " ")))
	}
}
