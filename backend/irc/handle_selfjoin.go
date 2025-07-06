package irc

import (
	"fmt"
	"github.com/ergochat/irc-go/ircmsg"
	"log/slog"
)

func HandleSelfJoin(
	timestampFormat string,
	setPendingUpdate func(),
	currentNick func() string,
	getChannelByName func(string) (*Channel, error),
	addChannel func(string) *Channel,
	hasCapability func(string) bool,
	sendRaw func(string),
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
		channel.AddMessage(NewEvent(EventJoin, timestampFormat, true, "You have joined "+channel.GetName()))
	}
}
