package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"log/slog"
	"strings"
)

func HandleNamesReply(
	setPendingUpdate setPendingUpdate,
	getChannelByName getChannelByName,
	getModePrefixes getModePrefixes,
) func(message ircmsg.Message) {
	return func(message ircmsg.Message) {
		stripChannelPrefixes := func(name string) (string, string) {
			prefixes := getModePrefixes()
			nickname := strings.TrimLeft(name, prefixes[1])
			modes := name[:len(name)-len(nickname)]
			return modes, nickname
		}
		defer setPendingUpdate()
		channel, err := getChannelByName(message.Params[2])
		if err != nil {
			slog.Debug("Names reply for unknown channel", "channel", message.Params[2])
			return
		}
		names := strings.Split(message.Params[3], " ")
		for i := range names {
			if names[i] == "" {
				continue
			}
			modes, nickname := stripChannelPrefixes(names[i])

			existingUsers := channel.GetUsers()
			userExists := false

			for j := range existingUsers {
				if existingUsers[j].nickname == nickname {
					existingUsers[j].modes = modes
					userExists = true
					break
				}
			}
			if !userExists {
				channel.AddUser(NewUser(nickname, modes))
			}
		}
	}
}
