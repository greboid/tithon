package irc

import (
	"encoding/json"
	"log"
	"net/http"
)

type SocketAction struct {
	Action  string          `json:"action"`
	Message json.RawMessage `json:"message"`
}

type SocketInit struct {
	Since int `json:"since"`
}

type SocketSendChannelMessage struct {
	Network string `json:"network"`
	Channel string `json:"channel"`
	Message string `json:"message"`
}

type SocketAddNetwork struct {
	Name     string `json:"name"`
	Server   string `json:"server"`
	TLS      bool   `json:"tls"`
	Nickname string `json:"nickname"`
	Realname string `json:"realname"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type SocketJoinChannel struct {
	Network string `json:"network"`
	Channel string `json:"channel"`
}

func SocketHandler(client *Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := client.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() {
			_ = conn.Close()
			client.RemoveClient(conn)
		}()
		client.AddClient(conn)
		for {
			actionMessage := &SocketAction{}
			err := conn.ReadJSON(actionMessage)
			if err != nil {
				log.Printf("Unable to parse incoming message: %s", err)
				break
			}
			switch actionMessage.Action {
			case "INIT":
				message := &SocketInit{}
				err := json.Unmarshal(actionMessage.Message, message)
				if err != nil {
					log.Printf("Unable to parse init message: %s", err.Error())
					break
				}
				client.InitClient(conn, message.Since)
				break
			case "SENDCHANMESSAGE":
				message := &SocketSendChannelMessage{}
				err := json.Unmarshal(actionMessage.Message, message)
				if err != nil {
					log.Printf("Unable to parse send chan message: %s", err.Error())
					break
				}
				client.SendMessage(message.Network, message.Channel, message.Message)
				break
			case "ADDNETWORK":
				message := &SocketAddNetwork{}
				err := json.Unmarshal(actionMessage.Message, message)
				if err != nil {
					log.Printf("Unable to parse message add network: %s", err.Error())
				}
				client.addNetwork(&Network{
					Name: message.Name,
					Profile: &Profile{
						Nickname: message.Nickname,
						Username: message.Username,
						Password: message.Password,
						RealName: message.Realname,
					},
					Hostname: message.Server,
					UseTLS:   message.TLS,
				})
			case "JOINCHANNEL":
				message := &SocketJoinChannel{}
				err := json.Unmarshal(actionMessage.Message, message)
				if err != nil {
					log.Printf("Unable to parse message join channel: %s", err.Error())
				}
				client.joinChannel(message.Network, message.Channel)
			}
		}
	}
}
