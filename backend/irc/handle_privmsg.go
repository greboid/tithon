package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"log/slog"
	"strings"
)

func HandlePrivMsg(
	timestampFormat string,
	setPendingUpdate func(),
	isValidChannel func(string) bool,
	getChannelByName func(string) (*Channel, error),
	currentNick func() string,
	getServerName func() string,
	checkAndNotify func(string, string, string, string) bool,
	getQueryByName func(string) (*Query, error),
	addQuery func(string) *Query,
) func(message ircmsg.Message) {
	return func(message ircmsg.Message) {
		if isCTCP(strings.Join(message.Params[1:], " ")) {
			return
		}
		defer setPendingUpdate()
		if isValidChannel(message.Params[0]) {
			channel, err := getChannelByName(message.Params[0])
			if err != nil {
				slog.Warn("Message for unknown channel", "message", message)
				return
			}
			msg := NewMessage(timestampFormat, message.Nick() == currentNick(), message.Nick(), strings.Join(message.Params[1:], " "), message.AllTags(), currentNick())
			if msg.tags["chathistory"] != "true" && !msg.IsMe() {
				checkAndNotify(getServerName(), channel.GetName(), msg.GetNickname(), msg.GetPlainDisplayMessage())
			}
			channel.AddMessage(msg)
		} else if strings.EqualFold(message.Params[0], currentNick()) {
			pm, err := getQueryByName(message.Nick())
			if err != nil {
				pm = addQuery(message.Nick())
			}

			msg := NewMessage(timestampFormat, message.Nick() == currentNick(), message.Nick(), strings.Join(message.Params[1:], " "), message.AllTags(), currentNick())
			if msg.tags["chathistory"] != "true" && !msg.IsMe() {
				checkAndNotify(getServerName(), pm.GetName(), msg.GetNickname(), msg.GetPlainDisplayMessage())
			}
			pm.AddMessage(msg)
		} else if message.Nick() == currentNick() {
			pm, err := getQueryByName(message.Params[0])
			if err != nil {
				pm = addQuery(message.Nick())
			}
			msg := NewMessage(timestampFormat, message.Nick() == currentNick(), message.Nick(), strings.Join(message.Params[1:], " "), message.AllTags(), currentNick())
			pm.AddMessage(msg)
		} else {
			slog.Warn("Unsupported message target", "message", message)
		}
	}
}
