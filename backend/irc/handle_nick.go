package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
)

func HandleNick(
	timestampFormat string,
	setPendingUpdate func(),
	currentNick func() string,
	addMessage func(*Message),
	getChannels func() []*Channel,
) func(message ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		if message.Nick() == currentNick() {
			addMessage(NewEvent(EventNick, timestampFormat, true, "Your nickname changed to "+message.Params[0]))
		}
		channels := getChannels()
		for i := range channels {
			users := channels[i].GetUsers()
			for j := range users {
				if users[j].nickname == message.Nick() {
					channels[i].AddMessage(NewEvent(EventNick, timestampFormat, false, message.Nick()+" is now known as "+message.Params[0]))
					users[j].nickname = message.Params[0]
				}
			}
		}
	}
}
