package events

type ServerAdded struct {
	Server Server  `json:"server"`
	Time   IRCTime `json:"time"`
}

type ServerUpdated struct {
	Server Server  `json:"server"`
	Time   IRCTime `json:"time"`
}

type ServerConnected struct {
	Server Server  `json:"server"`
	Time   IRCTime `json:"time"`
}

type ServerDisconnected struct {
	Server Server  `json:"server"`
	Time   IRCTime `json:"time"`
}

type ServerConnectionnError struct {
	Server Server  `json:"server"`
	Error  string  `json:"error"`
	Time   IRCTime `json:"time"`
}

type ServerMessageReceived struct {
	Message ServerMessage `json:"message"`
	Server  Server        `json:"server"`
}

type ServerMessageSent struct {
	Message ServerMessage `json:"message"`
	Server  Server        `json:"server"`
}

type DirectMessageReceived struct {
	Message DirectMessage `json:"message"`
	Server  Server        `json:"server"`
}

type DirectMessageSent struct {
	Message ServerMessage `json:"message"`
	Server  Server        `json:"server"`
}
