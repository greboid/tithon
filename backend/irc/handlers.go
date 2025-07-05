package irc

import (
	"github.com/ergochat/irc-go/ircevent"
	"regexp"
)

func AddCallbacks(linkRegex *regexp.Regexp, connection *Server, updateTrigger UpdateTrigger,
	notificationManager NotificationManager, timestampFormat string) {
	connection.AddCallback("JOIN", HandleSelfJoin(linkRegex, timestampFormat,
		updateTrigger.SetPendingUpdate, connection.CurrentNick, connection.GetChannelByName,
		connection.AddChannel, connection.HasCapability, connection.SendRaw))
	connection.AddCallback("JOIN", HandleOtherJoin(linkRegex, timestampFormat,
		updateTrigger.SetPendingUpdate, connection.CurrentNick, connection.GetChannelByName))
	connection.AddCallback("PRIVMSG", HandlePrivMsg(linkRegex, timestampFormat,
		updateTrigger.SetPendingUpdate, connection.IsValidChannel, connection.GetChannelByName,
		connection.CurrentNick, connection.GetName, notificationManager.CheckAndNotify,
		connection.GetQueryByName, connection.AddQuery))
	connection.AddCallback("NOTICE", HandleNotice(linkRegex, timestampFormat,
		updateTrigger.SetPendingUpdate, connection.CurrentNick, connection.AddMessage,
		connection.IsValidChannel, connection.GetChannelByName, connection.GetQueryByName,
		connection.AddQuery))
	connection.AddCallback(ircevent.RPL_TOPIC, HandleRPLTopic(updateTrigger.SetPendingUpdate,
		connection.GetName, connection.GetChannels))
	connection.AddCallback("333", HandleRPLTopicWhoTime(updateTrigger.SetPendingUpdate,
		connection.GetName, connection.GetChannelByName))
	connection.AddCallback("TOPIC", HandleTopic(linkRegex, timestampFormat,
		updateTrigger.SetPendingUpdate, connection.GetChannelByName, connection.GetName, connection.CurrentNick))
	connection.AddConnectCallback(HandleConnected(linkRegex, timestampFormat,
		updateTrigger.SetPendingUpdate, connection.GetQueries, connection.GetName, connection.ISupport,
		connection.SetName, connection.GetChannels, connection.AddMessage))
	connection.AddDisconnectCallback(HandleDisconnected(linkRegex, timestampFormat,
		updateTrigger.SetPendingUpdate, connection.GetQueries, connection.GetName,
		connection.GetChannels, connection.AddMessage))
	connection.AddCallback("PART", HandlePart(linkRegex, timestampFormat,
		updateTrigger.SetPendingUpdate, connection.CurrentNick, connection.GetChannelByName,
		connection.RemoveChannel))
	connection.AddCallback("KICK", HandleKick(linkRegex, timestampFormat,
		updateTrigger.SetPendingUpdate, connection.CurrentNick, connection.GetChannelByName,
		connection.RemoveChannel, connection.AddMessage))
	connection.AddCallback(ircevent.RPL_NAMREPLY, HandleNamesReply(updateTrigger.SetPendingUpdate,
		connection.GetChannelByName, connection.GetModePrefixes))
	connection.AddCallback(ircevent.RPL_UMODEIS, HandleUserModeSet(linkRegex, timestampFormat,
		updateTrigger.SetPendingUpdate, connection.SetCurrentModes, connection.AddMessage))
	connection.AddCallback("ERROR", HandleError(linkRegex, timestampFormat,
		updateTrigger.SetPendingUpdate, connection.AddMessage))
	connection.AddCallback(ircevent.RPL_WHOISUSER, HandleWhois(linkRegex, timestampFormat, connection.AddMessage))
	connection.AddCallback(ircevent.RPL_WHOISCERTFP, HandleWhois(linkRegex, timestampFormat, connection.AddMessage))
	connection.AddCallback(ircevent.RPL_WHOISACCOUNT, HandleWhois(linkRegex, timestampFormat, connection.AddMessage))
	connection.AddCallback(ircevent.RPL_WHOISBOT, HandleWhois(linkRegex, timestampFormat, connection.AddMessage))
	connection.AddCallback(ircevent.RPL_WHOISACTUALLY, HandleWhois(linkRegex, timestampFormat, connection.AddMessage))
	connection.AddCallback(ircevent.RPL_WHOISCHANNELS, HandleWhois(linkRegex, timestampFormat, connection.AddMessage))
	connection.AddCallback(ircevent.RPL_WHOISIDLE, HandleWhois(linkRegex, timestampFormat, connection.AddMessage))
	connection.AddCallback(ircevent.RPL_WHOISMODES, HandleWhois(linkRegex, timestampFormat, connection.AddMessage))
	connection.AddCallback(ircevent.RPL_WHOISOPERATOR, HandleWhois(linkRegex, timestampFormat, connection.AddMessage))
	connection.AddCallback(ircevent.RPL_WHOISSECURE, HandleWhois(linkRegex, timestampFormat, connection.AddMessage))
	connection.AddCallback(ircevent.RPL_WHOISSERVER, HandleWhois(linkRegex, timestampFormat, connection.AddMessage))
	connection.AddCallback(ircevent.RPL_ENDOFWHOIS, HandleWhois(linkRegex, timestampFormat, connection.AddMessage))
	connection.AddCallback("MODE", HandleChannelModes(linkRegex, timestampFormat,
		connection.IsValidChannel, updateTrigger.SetPendingUpdate, connection.GetChannelByName,
		connection.GetModeNameForMode, connection.GetChannelModeType))
	connection.AddCallback("MODE", HandleUserModes(linkRegex, timestampFormat,
		connection.IsValidChannel, updateTrigger.SetPendingUpdate, connection.GetCurrentModes,
		connection.SetCurrentModes, connection.AddMessage))
	connection.AddCallback("QUIT", HandleQuit(linkRegex, timestampFormat,
		updateTrigger.SetPendingUpdate, connection.GetChannels))
	connection.AddCallback("NICK", HandleNick(linkRegex, timestampFormat,
		updateTrigger.SetPendingUpdate, connection.CurrentNick, connection.AddMessage, connection.GetChannels))
	connection.AddBatchCallback(HandleBatch())
	connection.AddCallback(ircevent.ERR_NICKNAMEINUSE, HandleNickInUse(linkRegex,
		timestampFormat, updateTrigger.SetPendingUpdate, connection.AddMessage))
	connection.AddCallback(ircevent.ERR_PASSWDMISMATCH, HandlePasswordMismatch(linkRegex,
		timestampFormat, updateTrigger.SetPendingUpdate, connection.AddMessage))
}
