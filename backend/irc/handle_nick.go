package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"regexp"
)

func HandleNick(
	linkRegex *regexp.Regexp,
	timestampFormat string,
	setPendingUpdate setPendingUpdate,
	currentNick currentNick,
	addMessage addMessage,
	getChannels getChannels,
) func(message ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		if message.Nick() == currentNick() {
			addMessage(NewEvent(linkRegex, EventNick, timestampFormat, true, "Your nickname changed to "+message.Params[0]))
		}
		channels := getChannels()
		for i := range channels {
			users := channels[i].GetUsers()
			for j := range users {
				if users[j].nickname == message.Nick() {
					channels[i].AddMessage(NewEvent(linkRegex, EventNick, timestampFormat, false, message.Nick()+" is now known as "+message.Params[0]))
					users[j].nickname = message.Params[0]
				}
			}
		}
	}
}
