package irc

import (
	"github.com/ergochat/irc-go/ircevent"
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHandleWhois(t *testing.T) {
	timestampFormat := "15:04:05"

	tests := []struct {
		name            string
		message         ircmsg.Message
		expectedMessage string
	}{
		{
			name: "RPL_WHOISUSER",
			message: ircmsg.Message{
				Command: ircevent.RPL_WHOISUSER,
				Params:  []string{"requestor", "testnick", "testuser", "example.com", "*", "Test User"},
			},
			expectedMessage: "WHOIS: testnick testuser example.com * Test User",
		},
		{
			name: "RPL_WHOISCERTFP",
			message: ircmsg.Message{
				Command: ircevent.RPL_WHOISCERTFP,
				Params:  []string{"requestor", "testnick", "has client certificate fingerprint", "ABC123DEF456"},
			},
			expectedMessage: "WHOIS: testnick has client certificate fingerprint ABC123DEF456",
		},
		{
			name: "RPL_WHOISACCOUNT",
			message: ircmsg.Message{
				Command: ircevent.RPL_WHOISACCOUNT,
				Params:  []string{"requestor", "testnick", "accountname", "is logged in as"},
			},
			expectedMessage: "WHOIS testnick is logged in as accountname",
		},
		{
			name: "RPL_WHOISBOT",
			message: ircmsg.Message{
				Command: ircevent.RPL_WHOISBOT,
				Params:  []string{"requestor", "botnick", "is a bot on ExampleNet"},
			},
			expectedMessage: "WHOIS: botnick is a bot on ExampleNet",
		},
		{
			name: "RPL_WHOISACTUALLY",
			message: ircmsg.Message{
				Command: ircevent.RPL_WHOISACTUALLY,
				Params:  []string{"requestor", "testnick", "actualhost.example.com", "is actually using host"},
			},
			expectedMessage: "WHOIS: requestor testnick actualhost.example.com is actually using host",
		},
		{
			name: "RPL_WHOISCHANNELS",
			message: ircmsg.Message{
				Command: ircevent.RPL_WHOISCHANNELS,
				Params:  []string{"requestor", "testnick", "@#channel1 +#channel2 #channel3"},
			},
			expectedMessage: "WHOIS: testnick @#channel1 +#channel2 #channel3",
		},
		{
			name: "RPL_WHOISIDLE",
			message: ircmsg.Message{
				Command: ircevent.RPL_WHOISIDLE,
				Params:  []string{"requestor", "testnick", "300", "1640995200", "seconds idle, signon time"},
			},
			expectedMessage: "WHOIS: testnick 300 1640995200 seconds idle, signon time",
		},
		{
			name: "RPL_WHOISMODES",
			message: ircmsg.Message{
				Command: ircevent.RPL_WHOISMODES,
				Params:  []string{"requestor", "testnick", "is using modes +iwx"},
			},
			expectedMessage: "WHOIS: requestor testnick is using modes +iwx",
		},
		{
			name: "RPL_WHOISOPERATOR",
			message: ircmsg.Message{
				Command: ircevent.RPL_WHOISOPERATOR,
				Params:  []string{"requestor", "opnick", "is an IRC operator"},
			},
			expectedMessage: "WHOIS: opnick is an IRC operator",
		},
		{
			name: "RPL_WHOISSECURE",
			message: ircmsg.Message{
				Command: ircevent.RPL_WHOISSECURE,
				Params:  []string{"requestor", "securenick", "is using a secure connection"},
			},
			expectedMessage: "WHOIS: securenick is using a secure connection",
		},
		{
			name: "RPL_WHOISSERVER",
			message: ircmsg.Message{
				Command: ircevent.RPL_WHOISSERVER,
				Params:  []string{"requestor", "testnick", "irc.example.com", "Example IRC Network"},
			},
			expectedMessage: "WHOIS: testnick irc.example.com Example IRC Network",
		},
		{
			name: "RPL_ENDOFWHOIS",
			message: ircmsg.Message{
				Command: ircevent.RPL_ENDOFWHOIS,
				Params:  []string{"requestor", "testnick", "End of WHOIS list"},
			},
			expectedMessage: "WHOIS END testnick",
		},
		{
			name: "Unknown WHOIS command (should not match)",
			message: ircmsg.Message{
				Command: "999",
				Params:  []string{"requestor", "testnick", "some unknown response"},
			},
			expectedMessage: "",
		},
		{
			name: "RPL_WHOISUSER with minimal params",
			message: ircmsg.Message{
				Command: ircevent.RPL_WHOISUSER,
				Params:  []string{"requestor", "nick"},
			},
			expectedMessage: "WHOIS: nick",
		},
		{
			name: "RPL_WHOISACCOUNT with minimal params",
			message: ircmsg.Message{
				Command: ircevent.RPL_WHOISACCOUNT,
				Params:  []string{"requestor", "nick", "account"},
			},
			expectedMessage: "WHOIS nick is logged in as account",
		},
		{
			name: "Empty params for RPL_WHOISUSER",
			message: ircmsg.Message{
				Command: ircevent.RPL_WHOISUSER,
				Params:  []string{},
			},
			expectedMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messagesAdded := []*Message{}

			addMessage := func(msg *Message) {
				messagesAdded = append(messagesAdded, msg)
			}

			handler := HandleWhois(timestampFormat, addMessage)
			handler(tt.message)

			if tt.expectedMessage == "" {
				assert.Len(t, messagesAdded, 0, "Should not add any messages for unknown commands")
			} else {
				assert.Len(t, messagesAdded, 1, "Should add exactly one message")
				assert.Equal(t, tt.expectedMessage, messagesAdded[0].GetMessage(), "Message text should match")
				assert.Equal(t, MessageType(Event), messagesAdded[0].GetType(), "Should be event message type")
			}
		})
	}
}
