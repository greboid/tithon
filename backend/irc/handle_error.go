package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"strings"
)

func HandleError(
	timestampFormat string,
	setPendingUpdate func(),
	addMessage func(*Message),
) func(ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		addMessage(NewError(timestampFormat, false, strings.Join(message.Params, " ")))
	}
}

func HandleNickInUse(
	timestampFormat string,
	setPendingUpdate func(),
	addMessage func(*Message),
) func(ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		addMessage(NewError(timestampFormat, false, "Nickname in use: "+message.Params[1]))
	}
}

func HandlePasswordMismatch(
	timestampFormat string,
	setPendingUpdate func(),
	addMessage func(*Message),
) func(ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		addMessage(NewError(timestampFormat, false, "Password Mismatch: "+strings.Join(message.Params, " ")))
	}
}
