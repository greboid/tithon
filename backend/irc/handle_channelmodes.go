package irc

import (
	"fmt"
	"github.com/ergochat/irc-go/ircmsg"
	"log/slog"
	"strings"
)

func HandleChannelModes(
	timestampFormat string,
	isValidChannel func(string) bool,
	setPendingUpdate func(),
	getChannelByName func(string) (*Channel, error),
	getModeNameForMode func(string) string,
	getChannelModeType func(string) rune,
) func(ircmsg.Message) {
	return func(message ircmsg.Message) {
		type modeChange struct {
			mode      string
			change    bool
			nickname  string
			parameter string
			modeType  rune
		}
		var handleUserPrivilegeMode func(change modeChange, channel *Channel, message ircmsg.Message)

		handleUserPrivilegeMode = func(change modeChange, channel *Channel, message ircmsg.Message) {
			mode := getModeNameForMode(change.mode)
			users := channel.GetUsers()
			for j := range users {
				if users[j].nickname == change.nickname {
					if change.change {
						users[j].modes += mode
					} else {
						users[j].modes = strings.Replace(users[j].modes, mode, "", -1)
					}
				}
			}

			channel.SortUsers()

			var modeStr string
			if change.change {
				modeStr = "+" + change.mode
			} else {
				modeStr = "-" + change.mode
			}

			channel.AddMessage(NewEvent(EventMode, timestampFormat, false,
				fmt.Sprintf("%s sets mode %s %s", message.Nick(), modeStr, change.nickname)))
		}
		if !isValidChannel(message.Params[0]) {
			return
		}
		var ops []modeChange
		var add = true
		param := 2

		for i := 0; i < len(message.Params[1]); i++ {
			switch message.Params[1][i] {
			case '+':
				add = true
			case '-':
				add = false
			default:
				modeChar := string(message.Params[1][i])
				modeType := getChannelModeType(modeChar)

				change := modeChange{
					mode:     modeChar,
					change:   add,
					modeType: modeType,
				}

				needsParam := false
				skipMode := false

				switch modeType {
				case 'P':
					if param < len(message.Params) {
						change.nickname = message.Params[param]
						needsParam = true
					} else {
						// Skip privilege modes that don't have a parameter
						skipMode = true
					}
				case 'A':
					if param < len(message.Params) {
						change.parameter = message.Params[param]
						needsParam = true
					} else {
						// Skip type A modes that don't have a parameter
						skipMode = true
					}
				case 'B':
					if param < len(message.Params) {
						change.parameter = message.Params[param]
						needsParam = true
					} else {
						// Skip type B modes that don't have a parameter
						skipMode = true
					}
				case 'C':
					if add && param < len(message.Params) {
						change.parameter = message.Params[param]
						needsParam = true
					} else if add {
						// Skip type C modes when setting and no parameter available
						skipMode = true
					}
				case 'D': // Boolean setting - never needs parameter
				}

				// Only add the mode change if we're not skipping it
				if !skipMode {
					ops = append(ops, change)
				}

				if needsParam {
					param++
				}
			}
		}

		// Get channel once for all operations
		channel, err := getChannelByName(message.Params[0])
		if err != nil {
			slog.Warn("Received mode for unknown channel", "channel", message.Params[0])
			return
		}

		for _, op := range ops {

			defer setPendingUpdate()
			switch op.modeType {
			case 'P':
				handleUserPrivilegeMode(op, channel, message)
			case 'A', 'B', 'C', 'D':
				channel.SetChannelMode(op.modeType, op.mode, op.parameter, op.change)

				var modeStr string
				if op.change {
					modeStr = "+" + op.mode
				} else {
					modeStr = "-" + op.mode
				}

				var paramStr string
				if op.parameter != "" {
					paramStr = " " + op.parameter
				}

				channel.AddMessage(NewEvent(EventMode, timestampFormat, false,
					fmt.Sprintf("%s sets mode %s%s", message.Nick(), modeStr, paramStr)))
			default:
				slog.Warn("Unknown mode type", "mode", op.mode, "type", op.modeType)
			}
		}
	}
}
