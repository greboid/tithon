package irc

import (
	"fmt"
	"github.com/ergochat/irc-go/ircevent"
	"github.com/ergochat/irc-go/ircmsg"
	"log/slog"
	"slices"
	"strings"
)

type Handler struct {
	connection *Connection
}

func (h *Handler) addCallbacks() {
	h.connection.connection.AddCallback("JOIN", h.handleJoin)
	h.connection.connection.AddCallback("PRIVMSG", h.handlePrivMsg)
	h.connection.connection.AddCallback("NOTICE", h.handleNotice)
	h.connection.connection.AddCallback(ircevent.RPL_TOPIC, h.handleRPLTopic)
	h.connection.connection.AddCallback("TOPIC", h.handleTopic)
	h.connection.connection.AddConnectCallback(h.handleConnected)
	h.connection.connection.AddCallback("PART", h.handlePart)
	h.connection.connection.AddCallback("KICK", h.handleKick)
	h.connection.connection.AddCallback(ircevent.RPL_NAMREPLY, h.handleNameReply)
	h.connection.connection.AddCallback(ircevent.RPL_UMODEIS, h.handleUserMode)
	h.connection.connection.AddCallback("ERROR", h.handleError)
	h.connection.connection.AddCallback(ircevent.ERR_NICKNAMEINUSE, func(message ircmsg.Message) {
		h.addEvent("Nickname (" + message.Params[1] + ") already in use")
	})
	h.connection.connection.AddCallback("NICK", h.handleNick)
	h.connection.connection.AddCallback("QUIT", h.handleQuit)
	h.connection.connection.AddCallback(ircevent.ERR_PASSWDMISMATCH, func(message ircmsg.Message) {
		h.addEvent("Password Mismatch: " + strings.Join(message.Params, " "))
	})
	h.connection.connection.AddCallback("MODE", h.handleMode)
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
	channel, err := h.connection.GetChannelByName(message.Params[0])
	if err != nil {
		slog.Warn("Topic for unknown channel", "message", message)
		return
	}
	topic := NewTopic(strings.Join(message.Params[1:], " "))
	slog.Debug("Setting topic", "server", h.connection.GetName(), "channel", channel.GetName(), "topic", topic)
	channel.SetTopic(topic)
	channel.AddMessage(NewMessage("", message.Nick()+" changed the topic: "+topic.GetTopic(), Event))
}

func (h *Handler) handleRPLTopic(message ircmsg.Message) {
	for _, channel := range h.connection.GetChannels() {
		if channel.name == message.Params[1] {
			topic := NewTopic(strings.Join(message.Params[2:], " "))
			channel.SetTopic(topic)
			slog.Debug("Setting topic", "server", h.connection.GetName(), "channel", channel.GetName(), "topic", topic)
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
		channel.AddMessage(mess)
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
		h.connection.connection.SendRaw(fmt.Sprintf("CHATHISTORY LATEST %s * 100", message.Params[0]))
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
	channel.users = slices.DeleteFunc(channel.users, func(user *User) bool {
		return user.nickname == message.Nick()
	})
	channel.AddMessage(NewMessage("", message.Source+" has parted "+channel.GetName(), Event))
}

func (h *Handler) handleKick(message ircmsg.Message) {
	channel, err := h.connection.GetChannelByName(message.Params[0])
	if err != nil {
		slog.Warn("Received kick for unknown channel", "channel", message.Params[0])
		return
	}
	if message.Params[1] == h.connection.CurrentNick() {
		h.connection.RemoveChannel(channel.id)
		h.connection.AddMessage(NewMessage("", message.Nick()+" has kicked you from "+message.Params[0]+" ("+strings.Join(message.Params[2:], " ")+")", Event))
		return
	}
	channel.users = slices.DeleteFunc(channel.users, func(user *User) bool {
		return user.nickname == message.Nick()
	})
	channel.AddMessage(NewMessage("", message.Source+" has kicked "+message.Params[1]+" from "+channel.GetName()+"("+strings.Join(message.Params[2:], " ")+")", Event))
}

func (h *Handler) handleOtherJoin(message ircmsg.Message) {
	channel, err := h.connection.GetChannelByName(message.Params[0])
	if err != nil {
		slog.Error("Error getting channel for join", "message", message)
		return
	}
	channel.users = append(channel.users, NewUser(message.Nick(), ""))
	channel.AddMessage(NewMessage("", message.Source+" has joined "+channel.GetName(), Event))
}

func (h *Handler) handleConnected(message ircmsg.Message) {
	h.connection.AddMessage(NewMessage("", fmt.Sprintf("Connected to %s", h.connection.connection.Server), Event))
	network := h.connection.connection.ISupport()["NETWORK"]
	if len(network) > 0 {
		h.connection.Window.SetName(network)
	}
}

func (h *Handler) handleNameReply(message ircmsg.Message) {
	channel, err := h.connection.GetChannelByName(message.Params[2])
	if err != nil {
		slog.Debug("Names reply for unknown channel", "channel", message.Params[2])
		return
	}
	names := strings.Split(message.Params[3], " ")
	for i := range names {
		modes, user := h.stripChannelPrefixes(names[i])
		channel.AddUser(NewUser(user, modes))
	}
}

func (h *Handler) stripChannelPrefixes(name string) (string, string) {
	prefixes := h.connection.GetModePrefixes()
	nickname := strings.TrimLeft(name, prefixes[1])
	modes := name[:len(name)-len(nickname)]
	return modes, nickname
}

func (h *Handler) handleUserMode(message ircmsg.Message) {
	h.connection.currentModes = message.Params[1]
	h.connection.AddMessage(NewMessage("", "Your modes changed: "+message.Params[1], Event))
}

func (h *Handler) handleError(message ircmsg.Message) {
	h.connection.AddMessage(NewMessage("", strings.Join(message.Params, " "), Event))
}

func (h *Handler) handleNotice(message ircmsg.Message) {
	var mess *Message
	if found, messageTime := message.GetTag("time"); found {
		mess = NewMessageWithTime(messageTime, message.Nick(), strings.Join(message.Params[1:], " "), Notice)
	} else {
		mess = NewMessage(message.Nick(), strings.Join(message.Params[1:], " "), Notice)
	}
	if strings.Contains(message.Source, ".") && !strings.Contains(message.Source, "@") {
		h.connection.AddMessage(mess)
	} else if h.isChannel(message.Params[0]) {
		channel, err := h.connection.GetChannelByName(message.Params[0])
		if err != nil {
			slog.Warn("Notice for unknown channel", "notice", message)
			return
		}
		channel.AddMessage(mess)
	} else {
		slog.Warn("Unsupported DN", "notice", message)
	}
}

func (h *Handler) addEvent(message string) {
	h.connection.AddMessage(NewMessage("", message, Event))
}

func (h *Handler) handleNick(message ircmsg.Message) {
	if message.Nick() == h.connection.CurrentNick() {
		newNick := message.Params[0]
		h.connection.AddMessage(NewMessage("", "Nickname changed: "+newNick, Event))
	}
	channels := h.connection.GetChannels()
	for i := range channels {
		users := channels[i].GetUsers()
		for j := range users {
			if users[j].nickname == message.Nick() {
				channels[i].AddMessage(NewMessage("", message.Nick()+" is now known as "+message.Params[0], Event))
				users[j].nickname = message.Params[0]
			}
		}
	}
}

func (h *Handler) handleQuit(message ircmsg.Message) {
	channels := h.connection.GetChannels()
	for i := range channels {
		changed := false
		users := channels[i].GetUsers()
		users = slices.DeleteFunc(users, func(user *User) bool {
			if user.nickname == message.Nick() {
				changed = true
				return true
			}
			return false
		})
		if changed {
			channels[i].SetUsers(users)
			nuh, _ := message.NUH()
			channels[i].AddMessage(NewMessage("", nuh.Canonical()+" has quit "+strings.Join(message.Params[1:], " "), Event))
		}
	}
}

func (h *Handler) handleMode(message ircmsg.Message) {
	channel, err := h.connection.GetChannelByName(message.Params[0])
	if err != nil {
		slog.Warn("Received mode for unknown channel", "channel", message.Params[0])
		return
	}
	// TODO: Need to check the modes are in prefixes or channel modes and act accordingly, rather than assume
	// all modes are user modes
	type modeChange struct {
		mode     string
		change   bool
		nickname string
	}
	var ops []modeChange
	var add bool
	param := 2
	for i := 0; i < len(message.Params[1]); i++ {
		switch message.Params[1][i] {
		case '+':
			add = true
		case '-':
			add = false
		default:

			ops = append(ops, modeChange{
				mode:     string(message.Params[1][i]),
				change:   add,
				nickname: message.Params[param],
			})
			param++
		}
	}
	for i := range ops {
		users := channel.GetUsers()
		for j := range users {
			if users[j].nickname == ops[i].nickname {
				mode := h.connection.GetModeNameForMode(ops[i].mode)
				if ops[i].change {
					users[j].modes += mode
				} else {
					users[j].modes = strings.Replace(users[j].modes, mode, "", -1)
				}
			}
		}
	}
	channel.SortUsers()
}
