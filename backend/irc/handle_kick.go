package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"log/slog"
	"regexp"
	"slices"
	"strings"
)

func HandleKick(
	linkRegex *regexp.Regexp,
	timestampFormat string,
	setPendingUpdate func(),
	currentNick func() string,
	getChannelByName func(string) (*Channel, error),
	removeChannel func(string),
	addMessage func(*Message),
) func(message ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		channel, err := getChannelByName(message.Params[0])
		if err != nil {
			slog.Warn("Received kick for unknown channel", "channel", message.Params[0])
			return
		}
		kickMessage := strings.Join(message.Params[2:], " ")
		if kickMessage != "" {
			kickMessage = " (" + kickMessage + ")"
		}
		if message.Params[1] == currentNick() {
			removeChannel(channel.id)
			addMessage(NewEvent(linkRegex, EventKick, timestampFormat, true, message.Source+" has kicked you from "+channel.GetName()+kickMessage))
			return
		}
		channel.users = slices.DeleteFunc(channel.users, func(user *User) bool {
			return user.nickname == message.Params[1]
		})
		channel.AddMessage(NewEvent(linkRegex, EventKick, timestampFormat, message.Nick() == currentNick(), message.Source+" has kicked "+message.Params[1]+" from "+channel.GetName()+kickMessage))
	}
}
