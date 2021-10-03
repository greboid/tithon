package irc

import (
	"github.com/gorilla/websocket"
)

type ClientConfig struct {
	Networks []*Network
}

type Client struct {
	Networks         []*Network
	config           *ClientConfig
	connectedClients []*websocket.Conn
	upgrader         websocket.Upgrader
}

func NewIRCClient() *Client {
	conf := &ClientConfig{}
	client := &Client{
		config:       conf,
		upgrader:     websocket.Upgrader{},
	}
	return client
}

func (c *Client) AddClient(conn *websocket.Conn) {
	existing := false
	for i := range c.connectedClients {
		if c.connectedClients[i] == conn {
			existing = true
			break
		}
	}
	if !existing {
		c.connectedClients = append(c.connectedClients, conn)
	}
}

func (c *Client) RemoveClient(conn *websocket.Conn) {
	for i, v := range c.connectedClients {
		if v == conn {
			c.connectedClients = append(c.connectedClients[:i], c.connectedClients[i+1:]...)
			break
		}
	}
}

func (c *Client) Init() {
	for _, network := range c.config.Networks {
		network.Connect(c)
	}
}

func (c *Client) InitClient(conn *websocket.Conn, since int) {
	c.sendServerList(conn)
}

func (c *Client) sendServerLists() {
	for _, conn := range c.connectedClients {
		c.sendServerList(conn)
	}
}

func (c *Client) sendServerList(conn *websocket.Conn) {
	serverlist := make(map[string][]string)
	for _, network := range c.config.Networks {
		serverlist[network.Name] = []string{}
		for _, channel := range network.Channels {
			serverlist[network.Name] = append(serverlist[network.Name], channel.Name)
		}
	}
	_ = conn.WriteJSON(map[string]interface{}{"serverlist": serverlist})
}

func (c *Client) sendChannelMessage(network *Network, channel *Channel, message ChannelMessage) {
	for _, conn := range c.connectedClients {
		_ = conn.WriteJSON(map[string]interface{}{"network": network.Name, "channel": channel.Name, "message": message})
	}
}

func (c *Client) SendMessage(network string, channel string, message string) {
	for _, net := range c.config.Networks {
		if net.Name == network {
			for _, chann := range net.Channels {
				if chann.Name == channel {
					_ = net.connection.Send("PRIVMSG", channel, message)
				}
			}
		}
	}
}

func (c *Client) socketSender() {

}
