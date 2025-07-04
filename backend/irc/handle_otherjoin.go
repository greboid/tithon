package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"log/slog"
	"regexp"
)

func HandleOtherJoin(
	linkRegex *regexp.Regexp,
	timestampFormat string,
	setPendingUpdate setPendingUpdate,
	currentNick currentNick,
	getChannelByName getChannelByName,
) func(message ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		if len(message.Params) == 0 {
			slog.Debug("Invalid join message")
			return
		}
		if message.Nick() == currentNick() {
			return
		}
		channel, err := getChannelByName(message.Params[0])
		if err != nil {
			slog.Error("Error getting channel for join", "message", message)
			return
		}
		channel.AddUser(NewUser(message.Nick(), ""))
		channel.AddMessage(NewEvent(linkRegex, EventJoin, timestampFormat, false, message.Source+" has joined "+channel.GetName()))
	}
}
