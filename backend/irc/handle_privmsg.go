package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"log/slog"
	"regexp"
	"strings"
)

func HandlePrivMsg(
	linkRegex *regexp.Regexp,
	timestampFormat string,
	setPendingUpdate setPendingUpdate,
	isValidChannel isValidChannel,
	getChannelByName getChannelByName,
	currentNick currentNick,
	getServerName getServerName,
	checkAndNotify checkAndNotify,
	getQueryByName getQueryByName,
	addQuery addQuery,
) func(message ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		if isValidChannel(message.Params[0]) {
			channel, err := getChannelByName(message.Params[0])
			if err != nil {
				slog.Warn("Message for unknown channel", "message", message)
				return
			}
			msg := NewMessage(linkRegex, timestampFormat, message.Nick() == currentNick(), message.Nick(), strings.Join(message.Params[1:], " "), message.AllTags(), currentNick())
			if msg.tags["chathistory"] != "true" && !msg.isMe() {
				checkAndNotify(getServerName(), channel.GetName(), msg.GetNickname(), msg.GetPlainDisplayMessage())
			}
			channel.AddMessage(msg)
		} else if strings.ToLower(message.Params[0]) == strings.ToLower(currentNick()) {
			pm, err := getQueryByName(message.Nick())
			if err != nil {
				pm = addQuery(message.Nick())
			}

			msg := NewMessage(linkRegex, timestampFormat, message.Nick() == currentNick(), message.Nick(), strings.Join(message.Params[1:], " "), message.AllTags(), currentNick())
			if msg.tags["chathistory"] != "true" && !msg.isMe() {
				checkAndNotify(getServerName(), pm.GetName(), msg.GetNickname(), msg.GetPlainDisplayMessage())
			}
			pm.AddMessage(msg)
		} else if message.Nick() == currentNick() {
			pm, err := getQueryByName(message.Params[0])
			if err != nil {
				pm = addQuery(message.Nick())
			}
			msg := NewMessage(linkRegex, timestampFormat, message.Nick() == currentNick(), message.Nick(), strings.Join(message.Params[1:], " "), message.AllTags(), currentNick())
			pm.AddMessage(msg)
		} else {
			slog.Warn("Unsupported message target", "message", message)
		}
	}
}
