package irc

import (
	"github.com/greboid/tithon/config"
	"log/slog"
	"maps"
	"slices"
	"strings"
)

type UpdateTrigger interface {
	SetPendingUpdate()
}

type ConnectionManager struct {
	connections         map[string]*Connection
	commandManager      *CommandManager
	updateTrigger       UpdateTrigger
	notificationManager *NotificationManager
	config              *config.Config
}

func NewConnectionManager(conf *config.Config) *ConnectionManager {
	return &ConnectionManager{
		connections:    map[string]*Connection{},
		commandManager: NewCommandManager(conf),
		config:         conf,
	}
}

func (cm *ConnectionManager) AddConnection(
	id string,
	hostname string,
	port int,
	tls bool,
	password string,
	sasllogin string,
	saslpassword string,
	profile *Profile,
	connect bool,
) string {
	connection := NewConnection(cm.config, id, hostname, port, tls, password, sasllogin, saslpassword, profile, cm.updateTrigger, cm.notificationManager)
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

func (cm *ConnectionManager) Load() {
	for _, server := range cm.config.Servers {
		cm.AddConnection(server.ID, server.Hostname, server.Port, server.TLS, server.Password, server.SASLLogin, server.SASLPassword, NewProfile(server.Profile.Nickname), false)
	}
}

func (cm *ConnectionManager) Save() {
	slog.Debug("Saving connections to config")
	servers := make([]config.Server, 0)
	for _, server := range cm.connections {
		servers = append(servers, config.Server{
			ID:           server.id,
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
	cm.config.Servers = servers
	slog.Debug("Saved connections to config")
}

func (cm *ConnectionManager) SetUpdateTrigger(ut UpdateTrigger) {
	cm.updateTrigger = ut
}

func (cm *ConnectionManager) SetNotificationManager(nm *NotificationManager) {
	cm.notificationManager = nm
}
