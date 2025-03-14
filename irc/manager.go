package irc

import (
	"errors"
	"github.com/greboid/ircclient/config"
	"gopkg.in/yaml.v3"
	"log/slog"
	"maps"
	"os"
	"slices"
	"strings"
)

type ConnectionManager struct {
	connections    map[string]*Connection
	commandManager *CommandManager
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections:    map[string]*Connection{},
		commandManager: NewCommandManager(),
	}
}

func (cm *ConnectionManager) AddConnection(hostname string, port int, tls bool, password string, sasllogin string, saslpassword string, profile *Profile) string {
	connection := NewConnection(hostname, port, tls, password, sasllogin, saslpassword, profile)
	cm.connections[connection.id] = connection
	return connection.id
}

func (cm *ConnectionManager) RemoveConnection(id string) {
	cm.connections[id].Disconnect()
	delete(cm.connections, id)
}

func (cm *ConnectionManager) GetConnections() []*Connection {
	connections := slices.Collect(maps.Values(cm.connections))
	slices.SortStableFunc(connections, func(a, b *Connection) int {
		return strings.Compare(strings.ToLower(a.GetName()), strings.ToLower(b.GetName()))
	})
	return connections
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
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Error("Unable to load config", "error", err)
		return err
	}

	err = yaml.Unmarshal(yamlData, conf)
	if err != nil {
		return err
	}
	for _, server := range conf.Servers {
		cm.AddConnection(server.Hostname, server.Port, server.TLS, server.Password, server.SASLLogin, server.SASLPassword, NewProfile(server.Profile.Nickname))
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
			Password:     server.password,
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
