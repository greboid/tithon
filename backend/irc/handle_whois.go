package irc

import (
	"github.com/ergochat/irc-go/ircevent"
	"github.com/ergochat/irc-go/ircmsg"
	"regexp"
	"strings"
)

func HandleWhois(
	linkRegex *regexp.Regexp,
	timestampFormat string,
	addMessage func(*Message),
) func(message ircmsg.Message) {
	return func(message ircmsg.Message) {
		switch message.Command {
		case ircevent.RPL_WHOISUSER:
			if len(message.Params) >= 1 {
				addMessage(NewEvent(linkRegex, EventWhois, timestampFormat, false, "WHOIS: "+strings.Join(message.Params[1:], " ")))
			}
		case ircevent.RPL_WHOISCERTFP:
			addMessage(NewEvent(linkRegex, EventWhois, timestampFormat, false, "WHOIS: "+strings.Join(message.Params[1:], " ")))
		case ircevent.RPL_WHOISACCOUNT:
			authMessage := strings.TrimSpace(strings.Join(message.Params[3:], " "))
			if authMessage != "" {
				authMessage = " " + authMessage + " "
			} else {
				authMessage = " is logged in as "
			}
			addMessage(NewEvent(linkRegex, EventWhois, timestampFormat, false, "WHOIS "+message.Params[1]+authMessage+message.Params[2]))
		case ircevent.RPL_WHOISBOT:
			addMessage(NewEvent(linkRegex, EventWhois, timestampFormat, false, "WHOIS: "+strings.Join(message.Params[1:], " ")))
		case ircevent.RPL_WHOISACTUALLY:
			addMessage(NewEvent(linkRegex, EventWhois, timestampFormat, false, "WHOIS: "+strings.Join(message.Params, " ")))
		case ircevent.RPL_WHOISCHANNELS:
			addMessage(NewEvent(linkRegex, EventWhois, timestampFormat, false, "WHOIS: "+strings.Join(message.Params[1:], " ")))
		case ircevent.RPL_WHOISIDLE:
			addMessage(NewEvent(linkRegex, EventWhois, timestampFormat, false, "WHOIS: "+strings.Join(message.Params[1:], " ")))
		case ircevent.RPL_WHOISMODES:
			addMessage(NewEvent(linkRegex, EventWhois, timestampFormat, false, "WHOIS: "+strings.Join(message.Params, " ")))
		case ircevent.RPL_WHOISOPERATOR:
			addMessage(NewEvent(linkRegex, EventWhois, timestampFormat, false, "WHOIS: "+strings.Join(message.Params[1:], " ")))
		case ircevent.RPL_WHOISSECURE:
			addMessage(NewEvent(linkRegex, EventWhois, timestampFormat, false, "WHOIS: "+strings.Join(message.Params[1:], " ")))
		case ircevent.RPL_WHOISSERVER:
			addMessage(NewEvent(linkRegex, EventWhois, timestampFormat, false, "WHOIS: "+strings.Join(message.Params[1:], " ")))
		case ircevent.RPL_ENDOFWHOIS:
			addMessage(NewEvent(linkRegex, EventWhois, timestampFormat, false, "WHOIS END "+message.Params[1]))
		}
	}
}
