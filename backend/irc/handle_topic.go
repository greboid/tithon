package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"log/slog"
	"regexp"
	"strings"
	"time"
)

func HandleTopic(
	linkRegex *regexp.Regexp,
	timestampFormat string,
	setPendingUpdate setPendingUpdate,
	getChannelByName getChannelByName,
	getServerName getServerName,
	currentNick currentNick,
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
			channel.AddMessage(NewEvent(linkRegex, EventTopic, timestampFormat, message.Nick() == currentNick(), message.Nick()+" unset the topic"))
		} else {
			channel.AddMessage(NewEvent(linkRegex, EventTopic, timestampFormat, message.Nick() == currentNick(), message.Nick()+" changed the topic: "+topic.GetTopic()))
		}
	}
}
