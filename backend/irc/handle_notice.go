package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"log/slog"
	"strings"
)

func HandleNotice(
	timestampFormat string,
	setPendingUpdate func(),
	currentNick func() string,
	addMessage func(*Message),
	isValidChannel func(string) bool,
	getChannelByName func(string) (*Channel, error),
	getQueryByName func(string) (*Query, error),
	addQuery func(string) *Query,
) func(ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		mess := NewNotice(timestampFormat, message.Nick() == currentNick(), message.Nick(), strings.Join(message.Params[1:], " "), nil, currentNick())
		if message.Source == "" || (strings.Contains(message.Source, ".") && !strings.Contains(message.Source, "@")) {
			addMessage(mess)
		} else if isValidChannel(message.Params[0]) {
			channel, err := getChannelByName(message.Params[0])
			if err != nil {
				slog.Warn("Notice for unknown channel", "notice", message)
				return
			}
			channel.AddMessage(mess)
		} else if message.Params[0] == currentNick() {
			pm, err := getQueryByName(message.Nick())
			if err != nil {
				pm = addQuery(message.Nick())
			}
			pm.AddMessage(mess)
		} else if strings.ToLower(message.Params[0]) == strings.ToLower(currentNick()) {
			pm, err := getQueryByName(message.Nick())
			if err != nil {
				pm = addQuery(message.Nick())
			}
			pm.AddMessage(mess)
		} else {
			slog.Warn("Unsupported notice target", "notice", message)
		}
	}
}
