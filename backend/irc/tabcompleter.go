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

func NewQueryTabCompleter(query *Query, commandManager *CommandManager) TabCompleter {
	return NewCommandTabCompleter(&NoopTabCompleter{}, commandManager)
}

func NewServerTabCompleter(server *Server, commandManager *CommandManager) TabCompleter {
	return NewCommandTabCompleter(&NoopTabCompleter{}, commandManager)
}

func NewChannelTabCompleter(channel userList, commandManager *CommandManager) TabCompleter {
	return NewCommandTabCompleter(&ChannelTabCompleter{
		channel:       channel,
		previousIndex: -1,
	}, commandManager)
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

type CommandTabCompleter struct {
	wrapped        TabCompleter
	commandManager *CommandManager
	lastCompletion int
	lastInput      string
	lastPosition   int
	original       string
	orpos          int
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

func NewCommandTabCompleter(wrapped TabCompleter, commandManager *CommandManager) TabCompleter {
	return &CommandTabCompleter{
		wrapped:        wrapped,
		commandManager: commandManager,
		lastCompletion: -1,
	}
}

func (c *CommandTabCompleter) Complete(input string, position int) (string, int) {
	if position > 0 && input[0] == '/' {
		commandEnd := strings.Index(input, " ")
		if commandEnd == -1 {
			commandEnd = len(input)
		}

		if position <= commandEnd {
			return c.completeCommand(input, position)
		}
	}

	return c.wrapped.Complete(input, position)
}

func (c *CommandTabCompleter) completeCommand(input string, position int) (string, int) {
	commandPart := input[1:position]
	if spaceIndex := strings.Index(commandPart, " "); spaceIndex != -1 {
		commandPart = commandPart[:spaceIndex]
	}

	var lastCompletion int
	if c.lastInput == input && c.lastPosition == position {
		commandPart = c.original
		position = c.orpos
		lastCompletion = c.lastCompletion
	} else {
		lastCompletion = -1
	}

	choices := c.getAllCommands()

	match, completion := c.completePrefixInList(commandPart, choices, lastCompletion)

	commandEnd := strings.Index(input, " ")
	if commandEnd == -1 {
		commandEnd = len(input)
	}

	newInput := "/" + match + input[commandEnd:]
	newPosition := len("/" + match)

	c.lastCompletion = completion
	c.lastInput = newInput
	c.lastPosition = newPosition
	c.orpos = position
	c.original = commandPart

	return newInput, newPosition
}

func (c *CommandTabCompleter) getAllCommands() []string {
	var commands []string

	if c.commandManager != nil {
		for _, cmd := range c.commandManager.commands {
			commands = append(commands, cmd.GetName())
			for _, alias := range cmd.GetAliases() {
				commands = append(commands, alias)
			}
		}
	}

	return commands
}

func (c *CommandTabCompleter) completePrefixInList(start string, choices []string, lastMatch int) (string, int) {
	for i := range choices {
		index := (i + lastMatch + 1) % len(choices)
		if strings.HasPrefix(strings.ToLower(choices[index]), strings.ToLower(start)) {
			return choices[index], index
		}
	}
	return start, 0
}
