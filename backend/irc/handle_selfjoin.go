package irc

import (
	"fmt"
	"github.com/ergochat/irc-go/ircmsg"
	"log/slog"
	"regexp"
)

func HandleSelfJoin(
	linkRegex *regexp.Regexp,
	timestampFormat string,
	setPendingUpdate setPendingUpdate,
	currentNick currentNick,
	getChannelByName getChannelByName,
	addChannel addChannel,
	hasCapability hasCapability,
	sendRaw sendRaw,
) func(message ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		if len(message.Params) == 0 {
			slog.Debug("Invalid join message")
			return
		}
		if message.Nick() != currentNick() {
			return
		}
		slog.Debug("Joining channel", "channel", message.Params[0])
		channel, err := getChannelByName(message.Params[0])
		if err != nil {
			channel = addChannel(message.Params[0])
			if hasCapability("draft/chathistory") {
				sendRaw(fmt.Sprintf("CHATHISTORY LATEST %s * 100", message.Params[0]))
			}
		}
		channel.AddMessage(NewEvent(linkRegex, EventJoin, timestampFormat, false, "You have joined "+channel.GetName()))
	}
}
