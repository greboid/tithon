package irc

import (
	"fmt"
	"github.com/ergochat/irc-go/ircevent"
	"github.com/ergochat/irc-go/ircmsg"
	"log/slog"
	"slices"
	"strings"
	"time"
)

type Handler struct {
	connection *Connection
}

func (h *Handler) addCallbacks() {
	h.connection.connection.AddCallback("JOIN", h.handleJoin)
	h.connection.connection.AddCallback("PRIVMSG", h.handlePrivMsg)
	h.connection.connection.AddCallback(ircevent.RPL_TOPIC, h.handleRPLTopic)
	h.connection.connection.AddCallback("TOPIC", h.handleTopic)
	h.connection.connection.AddConnectCallback(h.handleConnected)
	h.connection.connection.AddCallback("PART", h.handlePart)
	h.connection.connection.AddCallback(ircevent.RPL_NAMREPLY, h.handleNameReply)
	h.connection.connection.AddCallback(ircevent.RPL_UMODEIS, h.handleUserMode)
	h.connection.connection.AddCallback(ircevent.RPL_CHANNELMODEIS, func(message ircmsg.Message) {})
	h.connection.connection.AddCallback(ircevent.RPL_TOPICTIME, func(message ircmsg.Message) {})
	h.connection.connection.AddCallback(ircevent.RPL_CREATIONTIME, func(message ircmsg.Message) {})
	h.connection.connection.AddCallback(ircevent.RPL_NOTOPIC, func(message ircmsg.Message) {})
	h.connection.connection.AddCallback(ircevent.RPL_MOTDSTART, func(message ircmsg.Message) {})
	h.connection.connection.AddCallback(ircevent.RPL_MOTD, func(message ircmsg.Message) {})
	h.connection.connection.AddCallback(ircevent.RPL_ENDOFMOTD, func(message ircmsg.Message) {})
	h.connection.connection.AddCallback(ircevent.ERR_NOMOTD, func(message ircmsg.Message) {})
	h.connection.connection.AddCallback(ircevent.RPL_AWAY, func(message ircmsg.Message) {})
	h.connection.connection.AddCallback(ircevent.RPL_UNAWAY, func(message ircmsg.Message) {})
	h.connection.connection.AddCallback(ircevent.RPL_NOWAWAY, func(message ircmsg.Message) {})
	h.connection.connection.AddCallback("ERROR", h.handleError)
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
		mess = NewMessageWithTime(messageTime, message.Nick(), strings.Join(message.Params[1:], " "), Normal)
	} else {
		mess = NewMessage(message.Nick(), strings.Join(message.Params[1:], " "), Normal)
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
	if h.connection.HasCapability("draft/chathistory") {
		timestamp := time.Now().AddDate(0, 0, -1)
		h.connection.connection.SendRaw(fmt.Sprintf("CHATHISTORY LATEST %s timestamp=%s 100", message.Params[0], timestamp.Format(v3TimestampFormat)))
	}
}

func (h *Handler) handlePart(message ircmsg.Message) {
	channel, err := h.connection.GetChannelByName(message.Params[0])
	if err != nil {
		slog.Warn("Received part for unknown channel", "channel", message.Params[0])
		return
	}
	if message.Nick() == h.connection.CurrentNick() {
		h.connection.RemoveChannel(channel.id)
		return
	}
	slices.DeleteFunc(channel.users, func(user *User) bool {
		return user.nickname == message.Nick()
	})
	channel.messages = append(channel.messages, NewMessage(message.Nick(), "Parted the channel", Event))
}

func (h *Handler) handleOtherJoin(message ircmsg.Message) {
}

func (h *Handler) handleConnected(message ircmsg.Message) {
	h.connection.messages = append(h.connection.messages, NewMessage("", "Server connected", Event))
}

func (h *Handler) handleNameReply(message ircmsg.Message) {
	channel, err := h.connection.GetChannelByName(message.Params[2])
	if err != nil {
		slog.Debug("Names reply for unknown channel", "channel", message.Params[2])
		return
	}
	names := strings.Split(message.Params[3], " ")
	for i := range names {
		user := h.stripChannelPrefixes(names[i])
		channel.users = append(channel.users, NewUser(user))
	}
	slices.SortFunc(channel.users, func(a, b *User) int {
		return strings.Compare(a.nickname, b.nickname)
	})
}

func (h *Handler) stripChannelPrefixes(name string) string {
	prefixes := h.connection.GetModePrefixes()
	return strings.TrimLeft(name, prefixes[1])
}

func (h *Handler) handleUserMode(message ircmsg.Message) {
	h.connection.currentModes = message.Params[1]
}

func (h *Handler) handleError(message ircmsg.Message) {
	h.connection.messages = append(h.connection.messages, NewMessage("", strings.Join(message.Params, " "), Event))
}
