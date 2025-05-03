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
	lastPosition   int
	lastInput      string
	original       string
	orpos          int
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
	users := t.nicknamesToString(t.channel.GetUsers())
	lastSpace, nextSpace := t.surroundingSpacesIndexes(input, position)
	var partial string
	if t.lastInput == input && t.lastPosition == position {
		partial = t.original
		position = t.orpos
		input = t.lastInput
	} else {
		partial = input[lastSpace:nextSpace]
	}
	match := t.completePrefixInList(partial, users, t.lastCompletion)
	diff := len(match) - len(partial)
	sentence := input[:lastSpace] + match + input[nextSpace:]
	t.lastInput = sentence
	t.lastPosition = position + diff
	t.orpos = position
	t.lastCompletion = match
	t.original = partial
	return sentence, position + diff
}

func (t *ChannelTabCompleter) surroundingSpacesIndexes(input string, position int) (int, int) {
	lastSpace := strings.LastIndex(input[:position], " ")
	if lastSpace == -1 {
		lastSpace = 0
	} else {
		lastSpace++
	}
	nextSpace := strings.Index(input[lastSpace:], " ")
	if nextSpace == -1 {
		nextSpace = len(input)
	}
	return lastSpace, nextSpace
}

func (t *ChannelTabCompleter) nicknamesToString(nicknames []*User) []string {
	output := make([]string, 0)
	for i := range nicknames {
		output = append(output, nicknames[i].nickname)
	}
	return output
}

func (t *ChannelTabCompleter) completePrefixInList(start string, choices []string, lastMatch string) string {
	for i := range choices {
		if strings.HasPrefix(strings.ToLower(choices[i]), strings.ToLower(start)) && strings.ToLower(lastMatch) != strings.ToLower(choices[i]) {
			return choices[i]
		}
	}
	return start
}
