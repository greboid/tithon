package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"log/slog"
	"strings"
	"time"
)

func HandleTopic(
	timestampFormat string,
	setPendingUpdate func(),
	getChannelByName func(string) (*Channel, error),
	getServerName func() string,
	currentNick func() string,
) func(message ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		channel, err := getChannelByName(message.Params[0])
		if err != nil {
			slog.Warn("Topic for unknown channel", "message", message)
			return
		}
		newTopic := strings.Join(message.Params[1:], " ")
		topic := NewTopic(newTopic, message.Nick(), time.Now())
		slog.Debug("Setting topic", "server", getServerName(), "channel", channel.GetName(), "topic", topic)
		channel.SetTopic(topic)
		channel.SetTitle(topic.GetDisplayTopic())
		if newTopic == "" {
			channel.AddMessage(NewEvent(EventTopic, timestampFormat, message.Nick() == currentNick(), message.Nick()+" unset the topic"))
		} else {
			channel.AddMessage(NewEvent(EventTopic, timestampFormat, message.Nick() == currentNick(), message.Nick()+" changed the topic: "+topic.GetTopic()))
		}
	}
}
