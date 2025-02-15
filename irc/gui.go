package irc

type GUI interface {
	AddConnection(server string, useTLS bool, SASLLogin string, SASLPassword string, PreferredNick string) (bool, error)
}
