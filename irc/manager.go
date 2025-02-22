package irc

import (
	"fmt"
	"github.com/albinj12/unique-id"
	"github.com/ergochat/irc-go/ircevent"
	"github.com/greboid/ircclient/config"
	"gopkg.in/yaml.v3"
	"log/slog"
	"maps"
	"os"
	"slices"
)

type ConnectionManager struct {
	connections map[string]*Connection
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: map[string]*Connection{},
	}
}

func (cm *ConnectionManager) AddConnection(hostname string, port int, tls bool, sasllogin string, saslpassword string, profile *Profile) string {
	s, _ := uniqueid.Generateid("a", 5, "s")
	useSasl := len(sasllogin) > 0 && len(saslpassword) > 0

	connection := &Connection{
		id:                s,
		hostname:          hostname,
		port:              port,
		tls:               tls,
		saslLogin:         sasllogin,
		saslPassword:      saslpassword,
		preferredNickname: profile.nickname,
		connection: &ircevent.Connection{
			Server:       fmt.Sprintf("%s:%d", hostname, port),
			Nick:         profile.nickname,
			SASLLogin:    sasllogin,
			SASLPassword: saslpassword,
			QuitMessage:  " ",
			Version:      " ",
			UseTLS:       tls,
			UseSASL:      useSasl,
			EnableCTCP:   false,
			Debug:        true,
		},
		channels: map[string]*Channel{},
	}
	cm.connections[connection.id] = connection
	return connection.id
}

func (cm *ConnectionManager) RemoveConnection(id string) {
	cm.connections[id].Disconnect()
	delete(cm.connections, id)
}

func (cm *ConnectionManager) GetConnections() []*Connection {
	return slices.Collect(maps.Values(cm.connections))
}

func (cm *ConnectionManager) GetConnection(id string) *Connection {
	return cm.connections[id]
}

func (cm *ConnectionManager) Start() {
	for _, connection := range cm.connections {
		connection.Connect()
	}
}

func (cm *ConnectionManager) Stop() {
	cm.Save()
	for _, connection := range cm.connections {
		connection.Disconnect()
	}
}

func (cm *ConnectionManager) Load() error {
	slog.Info("Loading config")
	conf := &config.Config{}

	yamlData, err := os.ReadFile("./config.yaml")
	if err != nil {
		slog.Error("Unable to load config", "error", err)
		return err
	}

	err = yaml.Unmarshal(yamlData, conf)
	if err != nil {
		return err
	}
	for _, server := range conf.Servers {
		cm.AddConnection(server.Hostname, server.Port, server.TLS, server.SASLLogin, server.SASLPassword, NewProfile(server.Profile.Nickname))
	}
	return nil
}

func (cm *ConnectionManager) Save() {
	slog.Info("Saving config")
	conf := &config.Config{}
	for _, server := range cm.connections {
		conf.Servers = append(conf.Servers, config.Server{
			Hostname:     server.hostname,
			Port:         server.port,
			TLS:          server.tls,
			SASLLogin:    server.saslLogin,
			SASLPassword: server.saslPassword,
			Profile: config.Profile{
				Nickname: server.preferredNickname,
			},
		})
	}
	data, err := yaml.Marshal(conf)
	if err != nil {
		slog.Error("Unable to save config", "error", err)
	}
	err = os.WriteFile("./config.yaml", data, 0644)
	if err != nil {
		slog.Error("Unable to save config", "error", err)
	}
}
