package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"log/slog"
	"strings"
)

type Handler struct {
	connection *Connection
}

func (h *Handler) addCallbacks() {
	h.connection.connection.AddCallback("JOIN", h.handleJoin)
	h.connection.connection.AddCallback("PRIVMSG", h.handlePrivMsg)
	h.connection.connection.AddCallback("332", h.handleRPLTopic)
	h.connection.connection.AddCallback("TOPIC", h.handleTopic)
	h.connection.connection.AddConnectCallback(h.handleConnected)
	h.connection.connection.AddCallback("PART", h.handlePart)
}

func (h *Handler) isChannel(target string) bool {
	chanTypes := h.connection.connection.ISupport()["CHANTYPES"]
	if chanTypes == "" {
		chanTypes = "#"
	}
	for _, char := range chanTypes {
		if strings.HasPrefix(target, string(char)) {
			return true
		}
	}
	return false
}

func (h *Handler) handleTopic(message ircmsg.Message) {
	slog.Debug("Handling topic", "message", message)
	for _, channel := range h.connection.GetChannels() {
		if channel.name == message.Params[0] {
			topic := NewTopic(strings.Join(message.Params[1:], " "))
			slog.Debug("Setting topic", "server", h.connection.GetName(), "channel", channel.GetName(), "topic", topic)
			channel.SetTopic(topic)
			return
		}
	}
}

func (h *Handler) handleRPLTopic(message ircmsg.Message) {
	for _, channel := range h.connection.GetChannels() {
		if channel.name == message.Params[1] {
			topic := NewTopic(strings.Join(message.Params[2:], " "))
			slog.Debug("Setting topic", "server", h.connection.GetName(), "channel", channel.GetName(), "topic", topic)
			channel.SetTopic(topic)
			return
		}
	}
}

func (h *Handler) handlePrivMsg(message ircmsg.Message) {
	var mess *Message
	if found, messageTime := message.GetTag("time"); found {
		mess = NewMessageWithTime(messageTime, message.Nick(), strings.Join(message.Params[1:], " "))
	} else {
		mess = NewMessage(message.Nick(), strings.Join(message.Params[1:], " "))
	}
	if h.isChannel(message.Params[0]) {
		channel, err := h.connection.GetChannelByName(message.Params[0])
		if err != nil {
			slog.Warn("Message for unknown channel", "message", message)
			return
		}
		channel.messages = append(channel.messages, mess)
	} else {
		slog.Warn("Unsupported DM", "message", message)
	}
}

func (h *Handler) handleJoin(message ircmsg.Message) {
	if message.Nick() == h.connection.CurrentNick() {
		h.handleSelfJoin(message)
	} else {
		h.handleOtherJoin(message)
	}
}

func (h *Handler) handleSelfJoin(message ircmsg.Message) {
	slog.Debug("Joining channel", "channel", message.Params[0])
	h.connection.AddChannel(message.Params[0])
}

func (h *Handler) handlePart(message ircmsg.Message) {
	channel, err := h.connection.GetChannelByName(message.Params[0])
	if err != nil {
		slog.Warn("Received part for unknown channel", "channel", message.Params[0])
		return
	}
	h.connection.RemoveChannel(channel.id)
}

func (h *Handler) handleOtherJoin(message ircmsg.Message) {

}

func (h *Handler) handleConnected(message ircmsg.Message) {

}
