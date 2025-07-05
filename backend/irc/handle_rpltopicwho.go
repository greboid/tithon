package irc

import (
	"fmt"
	"github.com/ergochat/irc-go/ircmsg"
	"log/slog"
	"strconv"
	"time"
)

func HandleRPLTopicWhoTime(
	setPendingUpdate func(),
	getServerName func() string,
	getChannelByName func(string) (*Channel, error),
) func(message ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		if len(message.Params) < 4 {
			return
		}
		channelName := message.Params[1]
		setBy := message.Params[2]
		timestamp, err := strconv.ParseInt(message.Params[3], 10, 64)
		if err != nil {
			slog.Warn("Failed to parse topic timestamp", "timestamp", message.Params[3], "error", err)
			return
		}
		setTime := time.Unix(timestamp, 0)

		channel, err := getChannelByName(channelName)
		if err != nil {
			slog.Debug("Received topic for unknown channel")
			return
		}
		existingTopic := channel.GetTopic().GetTopic()
		updatedTopic := NewTopic(existingTopic, setBy, setTime)
		channel.SetTopic(updatedTopic)
		channel.SetTitle(updatedTopic.GetDisplayTopic())
		slog.Debug("Updated topic who", "server", getServerName(), "channel", channel.GetName(), "setBy", setBy, "setTime", setTime)
	}
}
