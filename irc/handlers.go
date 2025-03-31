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

type channelHandler interface {
	IsChannel(target string) bool
	GetChannelByName(string) (*Channel, error)
	GetChannels() []*Channel
	AddChannel(name string) *Channel
	RemoveChannel(s string)
}

type callbackHandler interface {
	AddConnectCallback(callback func(message ircmsg.Message))
	AddCallback(command string, callback func(ircmsg.Message))
}

type infoHandler interface {
	ISupport(value string) string
	CurrentNick() string
	GetName() string
	SetName(string)
	GetHostname() string
	HasCapability(name string) bool
}

type modeHandler interface {
	GetModeNameForMode(mode string) string
	GetModePrefixes() []string
	GetCurrentModes() string
	SetCurrentModes(modes string)
}

type messageHandler interface {
	AddMessage(message *Message)
	SendRaw(message string)
}

type Handler struct {
	channelHandler  channelHandler
	callbackHandler callbackHandler
	infoHandler     infoHandler
	modeHandler     modeHandler
	messageHandler  messageHandler
}

func NewHandler(connection *Connection) *Handler {
	return &Handler{
		channelHandler:  connection,
		callbackHandler: connection,
		infoHandler:     connection,
		modeHandler:     connection,
		messageHandler:  connection,
	}
}

func (h *Handler) addCallbacks() {
	h.callbackHandler.AddCallback("JOIN", h.handleJoin)
	h.callbackHandler.AddCallback("PRIVMSG", h.handlePrivMsg)
	h.callbackHandler.AddCallback("NOTICE", h.handleNotice)
	h.callbackHandler.AddCallback(ircevent.RPL_TOPIC, h.handleRPLTopic)
	h.callbackHandler.AddCallback("TOPIC", h.handleTopic)
	h.callbackHandler.AddConnectCallback(h.handleConnected)
	h.callbackHandler.AddCallback("PART", h.handlePart)
	h.callbackHandler.AddCallback("KICK", h.handleKick)
	h.callbackHandler.AddCallback(ircevent.RPL_NAMREPLY, h.handleNameReply)
	h.callbackHandler.AddCallback(ircevent.RPL_UMODEIS, h.handleUserMode)
	h.callbackHandler.AddCallback("ERROR", h.handleError)
	h.callbackHandler.AddCallback(ircevent.ERR_NICKNAMEINUSE, func(message ircmsg.Message) {
		h.addEvent(GetTimeForMessage(message), "Nickname ("+message.Params[1]+") already in use")
	})
	h.callbackHandler.AddCallback("NICK", h.handleNick)
	h.callbackHandler.AddCallback("QUIT", h.handleQuit)
	h.callbackHandler.AddCallback(ircevent.ERR_PASSWDMISMATCH, func(message ircmsg.Message) {
		h.addEvent(GetTimeForMessage(message), "Password Mismatch: "+strings.Join(message.Params, " "))
	})
	h.callbackHandler.AddCallback("MODE", h.handleMode)
}

func (h *Handler) handleTopic(message ircmsg.Message) {
	channel, err := h.channelHandler.GetChannelByName(message.Params[0])
	if err != nil {
		slog.Warn("Topic for unknown channel", "message", message)
		return
	}
	newTopic := strings.Join(message.Params[1:], " ")
	topic := NewTopic(newTopic)
	slog.Debug("Setting topic", "server", h.infoHandler.GetName(), "channel", channel.GetName(), "topic", topic)
	channel.SetTopic(topic)
	channel.SetTitle(topic.GetTopic())
	if newTopic == "" {
		channel.AddMessage(NewEvent(GetTimeForMessage(message), message.Nick()+" unset the topic"))
	} else {
		channel.AddMessage(NewEvent(GetTimeForMessage(message), message.Nick()+" changed the topic: "+topic.GetTopic()))
	}
}

func (h *Handler) handleRPLTopic(message ircmsg.Message) {
	for _, channel := range h.channelHandler.GetChannels() {
		if channel.name == message.Params[1] {
			topic := NewTopic(strings.Join(message.Params[2:], " "))
			channel.SetTopic(topic)
			channel.SetTitle(topic.GetTopic())
			slog.Debug("Setting topic", "server", h.infoHandler.GetName(), "channel", channel.GetName(), "topic", topic)
			return
		}
	}
}

func (h *Handler) handlePrivMsg(message ircmsg.Message) {
	if h.channelHandler.IsChannel(message.Params[0]) {
		channel, err := h.channelHandler.GetChannelByName(message.Params[0])
		if err != nil {
			slog.Warn("Message for unknown channel", "message", message)
			return
		}
		channel.AddMessage(NewMessage(GetTimeForMessage(message), message.Nick(), strings.Join(message.Params[1:], " "), h.infoHandler.CurrentNick()))
	} else {
		slog.Warn("Unsupported DM", "message", message)
	}
}

func (h *Handler) handleJoin(message ircmsg.Message) {
	if message.Nick() == h.infoHandler.CurrentNick() {
		h.handleSelfJoin(message)
	} else {
		h.handleOtherJoin(message)
	}
}

func (h *Handler) handleSelfJoin(message ircmsg.Message) {
	slog.Debug("Joining channel", "channel", message.Params[0])
	h.channelHandler.AddChannel(message.Params[0])
	if h.infoHandler.HasCapability("draft/chathistory") {
		h.messageHandler.SendRaw(fmt.Sprintf("CHATHISTORY LATEST %s * 100", message.Params[0]))
	}
}

func (h *Handler) handlePart(message ircmsg.Message) {
	channel, err := h.channelHandler.GetChannelByName(message.Params[0])
	if err != nil {
		slog.Warn("Received part for unknown channel", "channel", message.Params[0])
		return
	}
	if message.Nick() == h.infoHandler.CurrentNick() {
		h.channelHandler.RemoveChannel(channel.id)
		return
	}
	channel.users = slices.DeleteFunc(channel.users, func(user *User) bool {
		return user.nickname == message.Nick()
	})
	channel.AddMessage(NewEvent(GetTimeForMessage(message), message.Source+" has parted "+channel.GetName()))
}

func (h *Handler) handleKick(message ircmsg.Message) {
	channel, err := h.channelHandler.GetChannelByName(message.Params[0])
	if err != nil {
		slog.Warn("Received kick for unknown channel", "channel", message.Params[0])
		return
	}
	if message.Params[1] == h.infoHandler.CurrentNick() {
		h.channelHandler.RemoveChannel(channel.id)
		h.messageHandler.AddMessage(NewEvent(GetTimeForMessage(message), message.Nick()+" has kicked you from "+message.Params[0]+" ("+strings.Join(message.Params[2:], " ")+")"))
		return
	}
	channel.users = slices.DeleteFunc(channel.users, func(user *User) bool {
		return user.nickname == message.Nick()
	})
	channel.AddMessage(NewEvent(GetTimeForMessage(message), message.Source+" has kicked "+message.Params[1]+" from "+channel.GetName()+"("+strings.Join(message.Params[2:], " ")+")"))
}

func (h *Handler) handleOtherJoin(message ircmsg.Message) {
	channel, err := h.channelHandler.GetChannelByName(message.Params[0])
	if err != nil {
		slog.Error("Error getting channel for join", "message", message)
		return
	}
	channel.users = append(channel.users, NewUser(message.Nick(), ""))
	channel.AddMessage(NewEvent(GetTimeForMessage(message), message.Source+" has joined "+channel.GetName()))
}

func (h *Handler) handleConnected(message ircmsg.Message) {
	h.messageHandler.AddMessage(NewEvent(GetTimeForMessage(message), fmt.Sprintf("Connected to %s", h.infoHandler.GetHostname())))
	network := h.infoHandler.ISupport("NETWORK")
	if len(network) > 0 {
		h.infoHandler.SetName(network)
	}
}

func (h *Handler) handleNameReply(message ircmsg.Message) {
	channel, err := h.channelHandler.GetChannelByName(message.Params[2])
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
	prefixes := h.modeHandler.GetModePrefixes()
	nickname := strings.TrimLeft(name, prefixes[1])
	modes := name[:len(name)-len(nickname)]
	return modes, nickname
}

func (h *Handler) handleUserMode(message ircmsg.Message) {
	h.modeHandler.SetCurrentModes(message.Params[1])
	h.messageHandler.AddMessage(NewEvent(GetTimeForMessage(message), "Your modes changed: "+message.Params[1]))
}

func (h *Handler) handleError(message ircmsg.Message) {
	h.messageHandler.AddMessage(NewEvent(GetTimeForMessage(message), strings.Join(message.Params, " ")))
}

func (h *Handler) handleNotice(message ircmsg.Message) {
	mess := NewNotice(GetTimeForMessage(message), message.Nick(), strings.Join(message.Params[1:], " "), h.infoHandler.CurrentNick())
	if strings.Contains(message.Source, ".") && !strings.Contains(message.Source, "@") {
		h.messageHandler.AddMessage(mess)
	} else if h.channelHandler.IsChannel(message.Params[0]) {
		channel, err := h.channelHandler.GetChannelByName(message.Params[0])
		if err != nil {
			slog.Warn("Notice for unknown channel", "notice", message)
			return
		}
		channel.AddMessage(mess)
	} else {
		slog.Warn("Unsupported DN", "notice", message)
	}
}

func (h *Handler) addEvent(timestamp time.Time, message string) {
	h.messageHandler.AddMessage(NewEvent(timestamp, message))
}

func (h *Handler) handleNick(message ircmsg.Message) {
	if message.Nick() == h.infoHandler.CurrentNick() {
		newNick := message.Params[0]
		h.messageHandler.AddMessage(NewEvent(GetTimeForMessage(message), "Nickname changed: "+newNick))
	}
	channels := h.channelHandler.GetChannels()
	for i := range channels {
		users := channels[i].GetUsers()
		for j := range users {
			if users[j].nickname == message.Nick() {
				channels[i].AddMessage(NewEvent(GetTimeForMessage(message), message.Nick()+" is now known as "+message.Params[0]))
				users[j].nickname = message.Params[0]
			}
		}
	}
}

func (h *Handler) handleQuit(message ircmsg.Message) {
	channels := h.channelHandler.GetChannels()
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
			channels[i].AddMessage(NewEvent(GetTimeForMessage(message), nuh.Canonical()+" has quit "+strings.Join(message.Params[1:], " ")))
		}
	}
}

func (h *Handler) handleMode(message ircmsg.Message) {
	channel, err := h.channelHandler.GetChannelByName(message.Params[0])
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
				mode := h.modeHandler.GetModeNameForMode(ops[i].mode)
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
