package irc

type PrivateMessage struct {
	id        string
	user      *User
	messages  []*Message
	conection *Connection
}
