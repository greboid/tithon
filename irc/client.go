package irc

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/gorilla/websocket"
	"github.com/syndtr/goleveldb/leveldb"
)

type ClientConfig struct {
	Networks []*Network
}

type Client struct {
	Networks         []*Network
	database         *leveldb.DB
	connectedClients []*websocket.Conn
	upgrader         websocket.Upgrader
}

func NewIRCClient(databaseDirectory string) (*Client, error) {
	db, err := leveldb.OpenFile(databaseDirectory, nil)
	if err != nil {
		return nil, err
	}
	client := &Client{
		upgrader: websocket.Upgrader{},
		database: db,
	}
	return client, nil
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

func (c *Client) Start() {
	networksBytes, err := c.database.Get([]byte("networks"), nil)
	if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
		log.Printf("Unable to start: %s", err.Error())
		return
	}
	var networks []*Network
	err = json.Unmarshal(networksBytes, &networks)
	if err != nil {
		log.Printf("Unable to start: %s", err.Error())
		return
	}
	c.Networks = networks
	for _, network := range c.Networks {
		network.Connect(c)
	}
}

func (c *Client) Stop() error {
	defer func() {
		_ = c.database.Close()
	}()
	networks, err := json.Marshal(c.Networks)
	if err != nil {
		return err
	}
	err = c.database.Put([]byte("networks"), networks, nil)
	if err != nil {
		return err
	}
	return nil
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
	for _, network := range c.Networks {
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
	for _, net := range c.Networks {
		if net.Name == network {
			for _, chann := range net.Channels {
				if chann.Name == channel {
					_ = net.connection.Send("PRIVMSG", channel, message)
				}
			}
		}
	}
}

func (c *Client) addNetwork(network *Network) {
	c.Networks = append(c.Networks, network)
	c.sendServerLists()
	network.Connect(c)
}

func (c *Client) joinChannel(network string, channel string) {
	for _, net := range c.Networks {
		if net.Name == network{
			_ = net.connection.Join(channel)
		}
	}
}