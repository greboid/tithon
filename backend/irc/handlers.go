package irc

import (
	"fmt"
	"github.com/ergochat/irc-go/ircevent"
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/greboid/tithon/config"
	"log/slog"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

type channelHandler interface {
	IsValidChannel(target string) bool
	GetChannelByName(string) (*Channel, error)
	GetChannels() []*Channel
	AddChannel(name string) *Channel
	RemoveChannel(s string)
}

type queryHandler interface {
	GetQueries() []*Query
	GetQueryByName(name string) (*Query, error)
	AddQuery(name string) *Query
}

type callbackHandler interface {
	AddConnectCallback(callback func(message ircmsg.Message))
	AddDisconnectCallback(callback func(message ircmsg.Message))
	AddCallback(command string, callback func(ircmsg.Message))
	AddBatchCallback(callback func(*ircevent.Batch) bool)
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
	GetChannelModeType(mode string) rune
}

type messageHandler interface {
	AddMessage(message *Message)
	SendRaw(message string)
}

type modeChange struct {
	mode      string
	change    bool
	nickname  string
	parameter string
	modeType  rune
}

type Handler struct {
	channelHandler      channelHandler
	queryHandler        queryHandler
	callbackHandler     callbackHandler
	infoHandler         infoHandler
	modeHandler         modeHandler
	messageHandler      messageHandler
	updateTrigger       UpdateTrigger
	notificationManager NotificationManager
	conf                *config.Config
	batchMap            map[string]string
	linkRegex           *regexp.Regexp
}

func NewHandler(linkRegex *regexp.Regexp, connection ServerInterface, ut UpdateTrigger, nm NotificationManager, conf *config.Config) *Handler {
	return &Handler{
		channelHandler:      connection,
		queryHandler:        connection,
		callbackHandler:     connection,
		infoHandler:         connection,
		modeHandler:         connection,
		messageHandler:      connection,
		updateTrigger:       ut,
		notificationManager: nm,
		conf:                conf,
		batchMap:            make(map[string]string),
		linkRegex:           linkRegex,
	}
}

func (h *Handler) addCallbacks() {
	h.callbackHandler.AddCallback("JOIN", h.handleJoin)
	h.callbackHandler.AddCallback("PRIVMSG", h.handlePrivMsg)
	h.callbackHandler.AddCallback("NOTICE", h.handleNotice)
	h.callbackHandler.AddCallback(ircevent.RPL_TOPIC, h.handleRPLTopic)
	h.callbackHandler.AddCallback("333", h.handleRPLTopicWhoTime)
	h.callbackHandler.AddCallback("TOPIC", h.handleTopic)
	h.callbackHandler.AddConnectCallback(h.handleConnected)
	h.callbackHandler.AddDisconnectCallback(h.handleDisconnected)
	h.callbackHandler.AddCallback("PART", h.handlePart)
	h.callbackHandler.AddCallback("KICK", h.handleKick)
	h.callbackHandler.AddCallback(ircevent.RPL_NAMREPLY, h.handleNameReply)
	h.callbackHandler.AddCallback(ircevent.RPL_UMODEIS, h.handleUserModeSet)
	h.callbackHandler.AddCallback("ERROR", h.handleError)
	h.callbackHandler.AddCallback(ircevent.ERR_NICKNAMEINUSE, func(message ircmsg.Message) {
		h.messageHandler.AddMessage(NewError(h.linkRegex, h.conf.UISettings.TimestampFormat, false, "Nickname in use: "+message.Params[1]))
	})
	h.callbackHandler.AddCallback("NICK", h.handleNick)
	h.callbackHandler.AddCallback("QUIT", h.handleQuit)
	h.callbackHandler.AddCallback(ircevent.ERR_PASSWDMISMATCH, func(message ircmsg.Message) {
		h.messageHandler.AddMessage(NewError(h.linkRegex, h.conf.UISettings.TimestampFormat, false, "Password Mismatch: "+strings.Join(message.Params, " ")))
	})
	h.callbackHandler.AddCallback("MODE", h.handleMode)
	h.callbackHandler.AddCallback(ircevent.RPL_WHOISUSER, func(message ircmsg.Message) {
		h.addEvent(EventWhois, false, "WHOIS: "+strings.Join(message.Params[1:], " "))
	})
	h.callbackHandler.AddCallback(ircevent.RPL_WHOISCERTFP, func(message ircmsg.Message) {
		h.addEvent(EventWhois, false, "WHOIS: "+strings.Join(message.Params[1:], " "))
	})
	h.callbackHandler.AddCallback(ircevent.RPL_WHOISACCOUNT, func(message ircmsg.Message) {
		h.addEvent(EventWhois, false, "WHOIS "+strings.Join(message.Params[2:], " ")+" "+message.Params[1])
	})
	h.callbackHandler.AddCallback(ircevent.RPL_WHOISBOT, func(message ircmsg.Message) {
		h.addEvent(EventWhois, false, "WHOIS: "+strings.Join(message.Params[1:], " "))
	})
	h.callbackHandler.AddCallback(ircevent.RPL_WHOISACTUALLY, func(message ircmsg.Message) {
		h.addEvent(EventWhois, false, "WHOIS: "+strings.Join(message.Params, " "))
	})
	h.callbackHandler.AddCallback(ircevent.RPL_WHOISCHANNELS, func(message ircmsg.Message) {
		h.addEvent(EventWhois, false, "WHOIS: "+strings.Join(message.Params[1:], " "))
	})
	h.callbackHandler.AddCallback(ircevent.RPL_WHOISIDLE, func(message ircmsg.Message) {
		h.addEvent(EventWhois, false, "WHOIS: "+strings.Join(message.Params[1:], " "))
	})
	h.callbackHandler.AddCallback(ircevent.RPL_WHOISMODES, func(message ircmsg.Message) {
		h.addEvent(EventWhois, false, "WHOIS: "+strings.Join(message.Params, " "))
	})
	h.callbackHandler.AddCallback(ircevent.RPL_WHOISOPERATOR, func(message ircmsg.Message) {
		h.addEvent(EventWhois, false, "WHOIS: "+strings.Join(message.Params[1:], " "))
	})
	h.callbackHandler.AddCallback(ircevent.RPL_WHOISSECURE, func(message ircmsg.Message) {
		h.addEvent(EventWhois, false, "WHOIS: "+strings.Join(message.Params[1:], " "))
	})
	h.callbackHandler.AddCallback(ircevent.RPL_WHOISSERVER, func(message ircmsg.Message) {
		h.addEvent(EventWhois, false, "WHOIS: "+strings.Join(message.Params[1:], " "))
	})
	h.callbackHandler.AddCallback(ircevent.RPL_ENDOFWHOIS, func(message ircmsg.Message) {
		h.addEvent(EventWhois, false, "WHOIS END "+message.Params[1])
	})
	h.callbackHandler.AddBatchCallback(h.handleBatch)
}

func (h *Handler) handleBatch(batch *ircevent.Batch) bool {
	if batch.Params[1] == "chathistory" {
		for i := range batch.Items {
			batch.Items[i].Message.SetTag("chathistory", "true")
		}
	}
	return false
}

func (h *Handler) handleTopic(message ircmsg.Message) {
	defer h.updateTrigger.SetPendingUpdate()
	channel, err := h.channelHandler.GetChannelByName(message.Params[0])
	if err != nil {
		slog.Warn("Topic for unknown channel", "message", message)
		return
	}
	newTopic := strings.Join(message.Params[1:], " ")
	topic := NewTopic(newTopic, message.Nick(), time.Now())
	slog.Debug("Setting topic", "server", h.infoHandler.GetName(), "channel", channel.GetName(), "topic", topic)
	channel.SetTopic(topic)
	channel.SetTitle(topic.GetDisplayTopic())
	if newTopic == "" {
		channel.AddMessage(NewEvent(h.linkRegex, EventTopic, h.conf.UISettings.TimestampFormat, h.isMsgMe(message), message.Nick()+" unset the topic"))
	} else {
		channel.AddMessage(NewEvent(h.linkRegex, EventTopic, h.conf.UISettings.TimestampFormat, h.isMsgMe(message), message.Nick()+" changed the topic: "+topic.GetTopic()))
	}
}

func (h *Handler) handleRPLTopic(message ircmsg.Message) {
	defer h.updateTrigger.SetPendingUpdate()
	for _, channel := range h.channelHandler.GetChannels() {
		if channel.name == message.Params[1] {
			topic := NewTopic(strings.Join(message.Params[2:], " "), "", time.Time{})
			channel.SetTopic(topic)
			channel.SetTitle(topic.GetDisplayTopic())
			slog.Debug("Setting topic", "server", h.infoHandler.GetName(), "channel", channel.GetName(), "topic", topic)
			return
		}
	}
}

func (h *Handler) handleRPLTopicWhoTime(message ircmsg.Message) {
	defer h.updateTrigger.SetPendingUpdate()
	if len(message.Params) < 4 {
		return
	}
	channelName := message.Params[1]
	setBy := message.Params[2]
	timestamp, err := strconv.ParseInt(message.Params[3], 10, 64)
	if err != nil {
		slog.Warn("Failed to parse topic timestamp", "timestamp", message.Params[3], "error", err)
		return
	}
	setTime := time.Unix(timestamp, 0)

	for _, channel := range h.channelHandler.GetChannels() {
		if channel.name == channelName {
			if channel.GetTopic() != nil {
				existingTopic := channel.GetTopic().GetTopic()
				updatedTopic := NewTopic(existingTopic, setBy, setTime)
				channel.SetTopic(updatedTopic)
				channel.SetTitle(updatedTopic.GetDisplayTopic())
				slog.Debug("Updated topic who", "server", h.infoHandler.GetName(), "channel", channel.GetName(), "setBy", setBy, "setTime", setTime)
			}
			return
		}
	}
}

func (h *Handler) handlePrivMsg(message ircmsg.Message) {
	defer h.updateTrigger.SetPendingUpdate()
	if h.channelHandler.IsValidChannel(message.Params[0]) {
		channel, err := h.channelHandler.GetChannelByName(message.Params[0])
		if err != nil {
			slog.Warn("Message for unknown channel", "message", message)
			return
		}
		msg := NewMessage(h.linkRegex, h.conf.UISettings.TimestampFormat, h.isMsgMe(message), message.Nick(), strings.Join(message.Params[1:], " "), message.AllTags(), h.infoHandler.CurrentNick())
		if msg.tags["chathistory"] != "true" && !msg.isMe() {
			h.notificationManager.CheckAndNotify(h.infoHandler.GetName(), channel.GetName(), msg.GetNickname(), msg.GetPlainDisplayMessage())
		}
		channel.AddMessage(msg)
	} else if strings.ToLower(message.Params[0]) == strings.ToLower(h.infoHandler.CurrentNick()) {
		pm, err := h.queryHandler.GetQueryByName(message.Nick())
		if err != nil {
			pm = h.queryHandler.AddQuery(message.Nick())
		}

		msg := NewMessage(h.linkRegex, h.conf.UISettings.TimestampFormat, h.isMsgMe(message), message.Nick(), strings.Join(message.Params[1:], " "), message.AllTags(), h.infoHandler.CurrentNick())
		if msg.tags["chathistory"] != "true" && !msg.isMe() {
			h.notificationManager.CheckAndNotify(h.infoHandler.GetName(), pm.GetName(), msg.GetNickname(), msg.GetPlainDisplayMessage())
		}
		pm.AddMessage(msg)
	} else if message.Nick() == h.infoHandler.CurrentNick() {
		pm, err := h.queryHandler.GetQueryByName(message.Params[0])
		if err != nil {
			pm = h.queryHandler.AddQuery(message.Nick())
		}
		msg := NewMessage(h.linkRegex, h.conf.UISettings.TimestampFormat, h.isMsgMe(message), message.Nick(), strings.Join(message.Params[1:], " "), message.AllTags(), h.infoHandler.CurrentNick())
		pm.AddMessage(msg)
	} else {
		slog.Warn("Unsupported message target", "message", message)
	}
}

func (h *Handler) handleJoin(message ircmsg.Message) {
	defer h.updateTrigger.SetPendingUpdate()
	if message.Nick() == h.infoHandler.CurrentNick() {
		h.handleSelfJoin(message)
	} else {
		h.handleOtherJoin(message)
	}
}

func (h *Handler) handleSelfJoin(message ircmsg.Message) {
	slog.Debug("Joining channel", "channel", message.Params[0])
	channel, err := h.channelHandler.GetChannelByName(message.Params[0])
	if err != nil {
		channel = h.channelHandler.AddChannel(message.Params[0])
		if h.infoHandler.HasCapability("draft/chathistory") {
			h.messageHandler.SendRaw(fmt.Sprintf("CHATHISTORY LATEST %s * 100", message.Params[0]))
		}
	}
	channel.AddMessage(NewEvent(h.linkRegex, EventJoin, h.conf.UISettings.TimestampFormat, false, "You have joined "+channel.GetName()))
}

func (h *Handler) handlePart(message ircmsg.Message) {
	defer h.updateTrigger.SetPendingUpdate()
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
	channel.AddMessage(NewEvent(h.linkRegex, EventJoin, h.conf.UISettings.TimestampFormat, false, message.Source+" has parted "+channel.GetName()))
}

func (h *Handler) handleKick(message ircmsg.Message) {
	defer h.updateTrigger.SetPendingUpdate()
	channel, err := h.channelHandler.GetChannelByName(message.Params[0])
	if err != nil {
		slog.Warn("Received kick for unknown channel", "channel", message.Params[0])
		return
	}
	if message.Params[1] == h.infoHandler.CurrentNick() {
		h.channelHandler.RemoveChannel(channel.id)
		h.addEvent(EventKick, true, message.Source+" has kicked you from "+channel.GetName()+"("+strings.Join(message.Params[2:], " ")+")")
		return
	}
	channel.users = slices.DeleteFunc(channel.users, func(user *User) bool {
		return user.nickname == message.Params[1]
	})
	channel.AddMessage(NewEvent(h.linkRegex, EventKick, h.conf.UISettings.TimestampFormat, h.isMsgMe(message), message.Source+" has kicked "+message.Params[1]+" from "+channel.GetName()+"("+strings.Join(message.Params[2:], " ")+")"))
}

func (h *Handler) handleOtherJoin(message ircmsg.Message) {
	channel, err := h.channelHandler.GetChannelByName(message.Params[0])
	if err != nil {
		slog.Error("Error getting channel for join", "message", message)
		return
	}
	channel.users = append(channel.users, NewUser(message.Nick(), ""))
	channel.AddMessage(NewEvent(h.linkRegex, EventJoin, h.conf.UISettings.TimestampFormat, false, message.Source+" has joined "+channel.GetName()))
}

func (h *Handler) handleDisconnected(message ircmsg.Message) {
	defer h.updateTrigger.SetPendingUpdate()
	disconnectMessage := fmt.Sprintf("Disconnected from %s: %s", h.infoHandler.GetHostname(), strings.Join(message.Params, " "))

	h.addEvent(EventDisconnected, false, disconnectMessage)
	for _, channel := range h.channelHandler.GetChannels() {
		channel.AddMessage(NewEvent(h.linkRegex, EventDisconnected, h.conf.UISettings.TimestampFormat, false, disconnectMessage))
	}
	for _, query := range h.queryHandler.GetQueries() {
		query.AddMessage(NewEvent(h.linkRegex, EventDisconnected, h.conf.UISettings.TimestampFormat, false, disconnectMessage))
	}
}

func (h *Handler) handleConnected(message ircmsg.Message) {
	defer h.updateTrigger.SetPendingUpdate()
	network := h.infoHandler.ISupport("NETWORK")
	if len(network) > 0 {
		h.infoHandler.SetName(network)
	}
	connectMessage := fmt.Sprintf("Connected to %s", h.infoHandler.GetHostname())

	h.addEvent(EventConnecting, false, connectMessage)
	for _, channel := range h.channelHandler.GetChannels() {
		channel.AddMessage(NewEvent(h.linkRegex, EventConnecting, h.conf.UISettings.TimestampFormat, false, connectMessage))
	}
	for _, query := range h.queryHandler.GetQueries() {
		query.AddMessage(NewEvent(h.linkRegex, EventConnecting, h.conf.UISettings.TimestampFormat, false, connectMessage))
	}
}

func (h *Handler) handleNameReply(message ircmsg.Message) {
	defer h.updateTrigger.SetPendingUpdate()
	channel, err := h.channelHandler.GetChannelByName(message.Params[2])
	if err != nil {
		slog.Debug("Names reply for unknown channel", "channel", message.Params[2])
		return
	}
	names := strings.Split(message.Params[3], " ")
	for i := range names {
		if names[i] == "" {
			continue
		}
		modes, nickname := h.stripChannelPrefixes(names[i])

		existingUsers := channel.GetUsers()
		userExists := false

		for j := range existingUsers {
			if existingUsers[j].nickname == nickname {
				existingUsers[j].modes = modes
				userExists = true
				break
			}
		}
		if !userExists {
			channel.AddUser(NewUser(nickname, modes))
		}
	}
}

func (h *Handler) stripChannelPrefixes(name string) (string, string) {
	prefixes := h.modeHandler.GetModePrefixes()
	nickname := strings.TrimLeft(name, prefixes[1])
	modes := name[:len(name)-len(nickname)]
	return modes, nickname
}

func (h *Handler) handleUserModeSet(message ircmsg.Message) {
	defer h.updateTrigger.SetPendingUpdate()
	h.modeHandler.SetCurrentModes(message.Params[1])
	h.addEvent(EventMode, false, "Your modes changed: "+message.Params[1])
}

func (h *Handler) handleError(message ircmsg.Message) {
	defer h.updateTrigger.SetPendingUpdate()
	h.messageHandler.AddMessage(NewError(h.linkRegex, h.conf.UISettings.TimestampFormat, false, strings.Join(message.Params, " ")))
}

func (h *Handler) handleNotice(message ircmsg.Message) {
	defer h.updateTrigger.SetPendingUpdate()
	mess := NewNotice(h.linkRegex, h.conf.UISettings.TimestampFormat, h.isMsgMe(message), message.Nick(), strings.Join(message.Params[1:], " "), nil, h.infoHandler.CurrentNick())
	if message.Source == "" || (strings.Contains(message.Source, ".") && !strings.Contains(message.Source, "@")) {
		h.messageHandler.AddMessage(mess)
	} else if h.channelHandler.IsValidChannel(message.Params[0]) {
		channel, err := h.channelHandler.GetChannelByName(message.Params[0])
		if err != nil {
			slog.Warn("Notice for unknown channel", "notice", message)
			return
		}
		channel.AddMessage(mess)
	} else if message.Params[0] == h.infoHandler.CurrentNick() {
		pm, err := h.queryHandler.GetQueryByName(message.Nick())
		if err != nil {
			pm = h.queryHandler.AddQuery(message.Nick())
		}
		pm.AddMessage(mess)
	} else if strings.ToLower(message.Params[0]) == strings.ToLower(h.infoHandler.CurrentNick()) {
		pm, err := h.queryHandler.GetQueryByName(message.Nick())
		if err != nil {
			pm = h.queryHandler.AddQuery(message.Nick())
		}
		pm.AddMessage(mess)
	} else {
		slog.Warn("Unsupported notice target", "notice", message)
	}
}

func (h *Handler) addEvent(eventType EventType, isMe bool, message string) {
	h.messageHandler.AddMessage(NewEvent(h.linkRegex, eventType, h.conf.UISettings.TimestampFormat, isMe, message))
}

func (h *Handler) handleNick(message ircmsg.Message) {
	defer h.updateTrigger.SetPendingUpdate()
	if h.isMsgMe(message) {
		h.addEvent(EventNick, true, "Your nickname changed to "+message.Params[0])
	}
	channels := h.channelHandler.GetChannels()
	for i := range channels {
		users := channels[i].GetUsers()
		for j := range users {
			if users[j].nickname == message.Nick() {
				channels[i].AddMessage(NewEvent(h.linkRegex, EventNick, h.conf.UISettings.TimestampFormat, false, message.Nick()+" is now known as "+message.Params[0]))
				users[j].nickname = message.Params[0]
			}
		}
	}
}

func (h *Handler) handleQuit(message ircmsg.Message) {
	defer h.updateTrigger.SetPendingUpdate()
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
			channels[i].AddMessage(NewEvent(h.linkRegex, EventNick, h.conf.UISettings.TimestampFormat, false, nuh.Canonical()+" has quit "+strings.Join(message.Params[1:], " ")))
		}
	}
}

func (h *Handler) handleMode(message ircmsg.Message) {
	defer h.updateTrigger.SetPendingUpdate()

	if h.channelHandler.IsValidChannel(message.Params[0]) {
		h.handleChannelModes(message)
	} else {
		h.handleUserMode(message)
	}
}

func (h *Handler) handleChannelModes(message ircmsg.Message) {
	var ops []modeChange
	var add = true
	param := 2

	for i := 0; i < len(message.Params[1]); i++ {
		switch message.Params[1][i] {
		case '+':
			add = true
		case '-':
			add = false
		default:
			modeChar := string(message.Params[1][i])
			modeType := h.modeHandler.GetChannelModeType(modeChar)

			change := modeChange{
				mode:     modeChar,
				change:   add,
				modeType: modeType,
			}

			needsParam := false
			skipMode := false

			switch modeType {
			case 'P':
				if param < len(message.Params) {
					change.nickname = message.Params[param]
					needsParam = true
				} else {
					// Skip privilege modes that don't have a parameter
					skipMode = true
				}
			case 'A':
				if param < len(message.Params) {
					change.parameter = message.Params[param]
					needsParam = true
				} else {
					// Skip type A modes that don't have a parameter
					skipMode = true
				}
			case 'B':
				if param < len(message.Params) {
					change.parameter = message.Params[param]
					needsParam = true
				} else {
					// Skip type B modes that don't have a parameter
					skipMode = true
				}
			case 'C':
				if add && param < len(message.Params) {
					change.parameter = message.Params[param]
					needsParam = true
				} else if add {
					// Skip type C modes when setting and no parameter available
					skipMode = true
				}
			case 'D': // Boolean setting - never needs parameter
			}

			// Only add the mode change if we're not skipping it
			if !skipMode {
				ops = append(ops, change)
			}

			if needsParam {
				param++
			}
		}
	}

	for _, op := range ops {
		h.applyChannelMode(op, message)
	}
}

func (h *Handler) applyChannelMode(change modeChange, message ircmsg.Message) {
	channel, err := h.channelHandler.GetChannelByName(message.Params[0])
	if err != nil {
		slog.Warn("Received mode for unknown channel", "channel", message.Params[0])
		return
	}

	switch change.modeType {
	case 'P':
		h.handleUserPrivilegeMode(change, channel, message)
	case 'A', 'B', 'C', 'D':
		channel.SetChannelMode(change.modeType, change.mode, change.parameter, change.change)

		var modeStr string
		if change.change {
			modeStr = "+" + change.mode
		} else {
			modeStr = "-" + change.mode
		}

		var paramStr string
		if change.parameter != "" {
			paramStr = " " + change.parameter
		}

		channel.AddMessage(NewEvent(h.linkRegex, EventMode, h.conf.UISettings.TimestampFormat, false,
			fmt.Sprintf("%s sets mode %s%s", message.Nick(), modeStr, paramStr)))
	default:
		slog.Warn("Unknown mode type", "mode", change.mode, "type", change.modeType)
	}
}

func (h *Handler) handleUserPrivilegeMode(change modeChange, channel *Channel, message ircmsg.Message) {
	mode := h.modeHandler.GetModeNameForMode(change.mode)

	users := channel.GetUsers()
	for j := range users {
		if users[j].nickname == change.nickname {
			if change.change {
				users[j].modes += mode
			} else {
				users[j].modes = strings.Replace(users[j].modes, mode, "", -1)
			}
		}
	}

	channel.SortUsers()

	var modeStr string
	if change.change {
		modeStr = "+" + change.mode
	} else {
		modeStr = "-" + change.mode
	}

	channel.AddMessage(NewEvent(h.linkRegex, EventMode, h.conf.UISettings.TimestampFormat, false,
		fmt.Sprintf("%s sets mode %s %s", message.Nick(), modeStr, change.nickname)))
}

func (h *Handler) handleUserMode(message ircmsg.Message) {
	defer h.updateTrigger.SetPendingUpdate()
	var add bool

	if len(message.Params) < 2 {
		slog.Warn("Invalid user mode message", "message", message)
		return
	}

	modeStr := message.Params[1]
	newModes := h.modeHandler.GetCurrentModes()

	for i := 0; i < len(modeStr); i++ {
		switch modeStr[i] {
		case '+':
			add = true
		case '-':
			add = false
		default:
			mode := string(modeStr[i])
			if add {
				if !strings.Contains(newModes, mode) {
					newModes += mode
				}
			} else {
				newModes = strings.Replace(newModes, mode, "", -1)
			}
		}
	}

	h.modeHandler.SetCurrentModes(newModes)
	var displayModeStr string
	if strings.HasPrefix(modeStr, "+") || strings.HasPrefix(modeStr, "-") {
		displayModeStr = modeStr
	} else {
		displayModeStr = "+" + modeStr
	}

	h.addEvent(EventMode, false, fmt.Sprintf("Your modes changed: %s", displayModeStr))
}

func (h *Handler) isMsgMe(message ircmsg.Message) bool {
	return message.Nick() == h.infoHandler.CurrentNick()
}
