package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"slices"
	"strings"
)

func HandleQuit(
	timestampFormat string,
	setPendingUpdate func(),
	getChannels func() []*Channel,
) func(ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		channels := getChannels()
		for i := range channels {
			changed := false
			users := channels[i].GetUsers()
			users = slices.DeleteFunc(users, func(user *User) bool {
				if user.nickname == message.Nick() {
					changed = true
					return true
				}
				return false
			})
			if changed {
				channels[i].SetUsers(users)
				nuh, _ := message.NUH()
				channels[i].AddMessage(NewEvent(EventNick, timestampFormat, false, nuh.Canonical()+" has quit "+strings.Join(message.Params[1:], " ")))
			}
		}
	}
}
