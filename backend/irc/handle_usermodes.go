package irc

import (
	"fmt"
	"github.com/ergochat/irc-go/ircmsg"
	"log/slog"
	"strings"
)

func HandleUserModes(
	timestampFormat string,
	isValidChannel func(string) bool,
	setPendingUpdate func(),
	getCurrentModes func() string,
	setCurrentModes func(string),
	addMessage func(*Message),
) func(ircmsg.Message) {
	return func(message ircmsg.Message) {
		if isValidChannel(message.Params[0]) {
			return
		}
		defer setPendingUpdate()
		var add bool = true

		if len(message.Params) < 2 {
			slog.Warn("Invalid user mode message", "message", message)
			return
		}

		modeStr := message.Params[1]
		newModes := getCurrentModes()

		for i := 0; i < len(modeStr); i++ {
			switch modeStr[i] {
			case '+':
				add = true
			case '-':
				add = false
			default:
				mode := string(modeStr[i])
				if add {
					if !strings.Contains(newModes, mode) {
						newModes += mode
					}
				} else {
					newModes = strings.Replace(newModes, mode, "", -1)
				}
			}
		}

		setCurrentModes(newModes)
		var displayModeStr string
		if strings.HasPrefix(modeStr, "+") || strings.HasPrefix(modeStr, "-") {
			displayModeStr = modeStr
		} else {
			displayModeStr = "+" + modeStr
		}

		addMessage(NewEvent(EventMode, timestampFormat, true, fmt.Sprintf("Your modes changed: %s", displayModeStr)))
	}
}
