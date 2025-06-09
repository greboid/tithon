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

type QueryTabCompleter struct{}

type ChannelTabCompleter struct {
	channel        userList
	lastCompletion int
	previousIndex  int
	lastPosition   int
	lastInput      string
	original       string
	orpos          int
}

func NewQueryTabCompleter(query *Query) TabCompleter {
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
	var lastCompletion int
	if t.lastInput == input && t.lastPosition == position {
		partial = t.original
		position = t.orpos
		input = t.lastInput
		lastCompletion = t.lastCompletion
	} else {
		partial = input[lastSpace:nextSpace]
		lastCompletion = -1
	}
	match, completion := t.completePrefixInList(partial, users, lastCompletion)
	diff := len(match) - len(partial)
	sentence := input[:lastSpace] + match + input[nextSpace:]
	t.lastCompletion = completion
	t.lastInput = sentence
	t.lastPosition = position + diff
	t.orpos = position
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

func (t *ChannelTabCompleter) completePrefixInList(start string, choices []string, lastMatch int) (string, int) {
	for i := range choices {
		index := (i + lastMatch + 1) % len(choices)
		if strings.HasPrefix(strings.ToLower(choices[index]), strings.ToLower(start)) {
			return choices[index], index
		}
	}
	return start, 0
}
