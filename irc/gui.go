package irc

import "newirc/events"

type GUI interface {
	AddConnection(server string, useTLS bool, SASLLogin string, SASLPassword string, PreferredNick string) (bool, error)
	UIReady()
	GetChannel(channel events.Channel)
	GetServer(server events.Server)
}
