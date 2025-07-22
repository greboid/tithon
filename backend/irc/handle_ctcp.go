package irc

import (
	"fmt"
	"github.com/ergochat/irc-go/ircmsg"
	"log/slog"
	"strings"
)

const CTCPDelimiter = "\x01"

type CTCPMessage struct {
	Command    string
	Parameters string
	IsQuery    bool
}

func isCTCP(text string) bool {
	return strings.HasPrefix(text, CTCPDelimiter) && strings.HasSuffix(text, CTCPDelimiter) && len(text) >= 8 && text[1:7] != "ACTION"
}

func ParseCTCPMessage(text string) (*CTCPMessage, bool) {
	if !isCTCP(text) {
		return nil, false
	}

	content := strings.Trim(text, CTCPDelimiter)
	if content == "" {
		return nil, false
	}

	parts := strings.SplitN(content, " ", 2)
	command := strings.ToUpper(parts[0])
	parameters := ""
	if len(parts) > 1 {
		parameters = parts[1]
	}

	if command != "VERSION" {
		return nil, false
	}

	return &CTCPMessage{
		Command:    command,
		Parameters: parameters,
		IsQuery:    true,
	}, true
}

func FormatCTCPReply(command, parameters string) string {
	if parameters == "" {
		return fmt.Sprintf("%s%s%s", CTCPDelimiter, command, CTCPDelimiter)
	}
	return fmt.Sprintf("%s%s %s%s", CTCPDelimiter, command, parameters, CTCPDelimiter)
}

func HandleCTCPQuery(
	timestampFormat string,
	setPendingUpdate func(),
	sendRaw func(string),
	addMessage func(*Message),
) func(ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		ctcp, is := ParseCTCPMessage(strings.Join(message.Params[1:], " "))
		if !is {
			return
		}
		if ctcp.Command == "VERSION" {
			sendRaw(fmt.Sprintf("NOTICE %s :%s", message.Nick(),
				FormatCTCPReply("VERSION", "Tithon")))
			addMessage(NewEvent(EventHelp, timestampFormat, false,
				fmt.Sprintf("CTCP %s query from %s", ctcp.Command, message.Nick())))
			return
		}
		slog.Debug("Unknown CTCP command", "command", ctcp.Command)
	}
}

func HandleCTCPReply(
	timestampFormat string,
	setPendingUpdate func(),
	addMessage func(*Message),
) func(ircmsg.Message) {
	return func(message ircmsg.Message) {
		defer setPendingUpdate()
		ctcp, is := ParseCTCPMessage(strings.Join(message.Params[1:], " "))
		if !is {
			return
		}

		if ctcp.Command == "VERSION" {
			addMessage(NewEvent(EventHelp, timestampFormat, false, fmt.Sprintf("CTCP VERSION reply from %s: %s", message.Nick(), ctcp.Parameters)))
			return
		}
		slog.Debug("Unknown CTCP reply", "command", ctcp.Command)
	}
}
