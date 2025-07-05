package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"log/slog"
	"regexp"
	"slices"
)

func HandlePart(
	linkRegex *regexp.Regexp,
	timestampFormat string,
	setPendingUpdate func(),
	currentNick func() string,
	getChannelByName func(string) (*Channel, error),
	removeChannel func(string),
) func(message ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		channel, err := getChannelByName(message.Params[0])
		if err != nil {
			slog.Warn("Received part for unknown channel", "channel", message.Params[0])
			return
		}
		if message.Nick() == currentNick() {
			removeChannel(channel.id)
			return
		}
		channel.users = slices.DeleteFunc(channel.users, func(user *User) bool {
			return user.nickname == message.Nick()
		})
		channel.AddMessage(NewEvent(linkRegex, EventJoin, timestampFormat, false, message.Source+" has parted "+channel.GetName()))
	}
}
