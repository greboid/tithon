package irc

import (
	"strings"
)

type userList interface {
	GetUsers() []*User
}

type TabCompleter interface {
	Complete(input string, position int) (string, int)
}

type NoopTabCompleter struct{}

type PMTabCompleter struct{}

type ChannelTabCompleter struct {
	channel        userList
	lastCompletion string
	previousIndex  int
}

func NewPrivateMessageTabCompleter(pm *PrivateMessage) TabCompleter {
	return &NoopTabCompleter{}
}

func NewConnectionTabCompleter(connection *Connection) TabCompleter {
	return &NoopTabCompleter{}
}

func NewChannelTabCompleter(channel userList) TabCompleter {
	return &ChannelTabCompleter{
		channel:       channel,
		previousIndex: -1,
	}
}

func (t *NoopTabCompleter) Complete(input string, position int) (string, int) {
	return input, position
}

func (t *ChannelTabCompleter) Complete(input string, position int) (string, int) {
	if position > len(input) {
		return input, position
	}
	lastSpace := strings.LastIndex(input[:position], " ")
	if lastSpace == -1 {
		lastSpace = 0
	} else {
		lastSpace++
	}
	users := t.channel.GetUsers()
	matched := false
	var output string
	var outputPos int
	for i := range users {
		if strings.HasPrefix(users[i].nickname, t.lastCompletion) {
			matched = true
			if len(input) == position {
				output = input[:lastSpace] + users[i].nickname + " "
				outputPos = len(input[:lastSpace]) + len(users[i].nickname)
				t.lastCompletion = users[i].nickname
			} else {
				oldLength := position - lastSpace
				output = input[:lastSpace] + users[i].nickname + " " + input[:oldLength]
				outputPos = len(input[:lastSpace]) + len(users[i].nickname) + 1
				t.lastCompletion = users[i].nickname
			}
			if t.previousIndex != -1 {
				if t.previousIndex <= i {
					continue
				}
			}
			t.previousIndex = i
			return output, outputPos
		}
	}
	if matched {
		return output, outputPos
	}
	t.previousIndex = -1
	return input, position
}
