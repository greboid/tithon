package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"log/slog"
	"strings"
	"time"
)

func HandleRPLTopic(
	setPendingUpdate func(),
	getServerName func() string,
	getChannels func() []*Channel,
) func(message ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		for _, channel := range getChannels() {
			if channel.name == message.Params[1] {
				topic := NewTopic(strings.Join(message.Params[2:], " "), "", time.Time{})
				channel.SetTopic(topic)
				channel.SetTitle(topic.GetDisplayTopic())
				slog.Debug("Setting topic", "server", getServerName(), "channel", channel.GetName(), "topic", topic)
				return
			}
		}
	}
}
