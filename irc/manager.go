package irc

import (
	"errors"
	"github.com/greboid/ircclient/config"
	"github.com/kirsle/configdir"
	"gopkg.in/yaml.v3"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type UpdateTrigger interface {
	SetPendingUpdate()
}

type ConnectionManager struct {
	connections    map[string]*Connection
	commandManager *CommandManager
	updateTrigger  UpdateTrigger
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections:    map[string]*Connection{},
		commandManager: NewCommandManager(),
	}
}

func (cm *ConnectionManager) AddConnection(
	hostname string,
	port int,
	tls bool,
	password string,
	sasllogin string,
	saslpassword string,
	profile *Profile,
	connect bool,
) string {
	connection := NewConnection(hostname, port, tls, password, sasllogin, saslpassword, profile, cm.updateTrigger)
	cm.connections[connection.id] = connection
	if connect {
		go func() {
			connection.Connect()
		}()
	}
	cm.updateTrigger.SetPendingUpdate()
	return connection.id
}

func (cm *ConnectionManager) RemoveConnection(id string) {
	cm.connections[id].Disconnect()
	delete(cm.connections, id)
	cm.updateTrigger.SetPendingUpdate()
}

func (cm *ConnectionManager) GetConnections() []*Connection {
	connections := slices.Collect(maps.Values(cm.connections))
	slices.SortStableFunc(connections, func(a, b *Connection) int {
		if a.GetName() == b.GetName() {
			return strings.Compare(a.GetID(), b.GetID())
		}
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
	configPath := configdir.LocalConfig("ircclient")
	err := configdir.MakePath(configPath)
	if err != nil {
		return err
	}

	slog.Info("Loading config")
	conf := &config.Config{}

	yamlData, err := os.ReadFile(filepath.Join(configPath, "config.yaml"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Error("Unable to load config", "error", err)
		return err
	}

	err = yaml.Unmarshal(yamlData, conf)
	if err != nil {
		return err
	}
	for _, server := range conf.Servers {
		cm.AddConnection(server.Hostname, server.Port, server.TLS, server.Password, server.SASLLogin, server.SASLPassword, NewProfile(server.Profile.Nickname), false)
	}
	return nil
}

func (cm *ConnectionManager) Save() {
	configPath := configdir.LocalConfig("ircclient")
	err := configdir.MakePath(configPath)
	if err != nil {
		return
	}

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
	err = os.WriteFile(filepath.Join(configPath, "config.yaml"), data, 0644)
	if err != nil {
		slog.Error("Unable to save config", "error", err)
	}
}

func (cm *ConnectionManager) SetUpdateTrigger(ut UpdateTrigger) {
	cm.updateTrigger = ut
}
