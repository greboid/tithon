package irc

import (
	"github.com/greboid/tithon/config"
	"maps"
	"regexp"
	"slices"
	"strings"
)

type UpdateTrigger interface {
	SetPendingUpdate()
}

type ServerManager struct {
	connections         map[string]ServerInterface
	commandManager      *CommandManager
	updateTrigger       UpdateTrigger
	notificationManager NotificationManager
	config              *config.Config
	linkRegex           *regexp.Regexp
}

func NewServerManager(linkRegex *regexp.Regexp, conf *config.Config, commandManager *CommandManager) *ServerManager {
	return &ServerManager{
		connections:    map[string]ServerInterface{},
		commandManager: commandManager,
		config:         conf,
		linkRegex:      linkRegex,
	}
}

func (cm *ServerManager) AddConnection(
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
	connection := NewServer(cm.linkRegex, cm.config, id, hostname, port, tls, password, sasllogin, saslpassword, profile, cm.updateTrigger, cm.notificationManager)
	cm.connections[connection.GetID()] = connection
	if connect {
		go func() {
			connection.Connect()
		}()
	}
	cm.updateTrigger.SetPendingUpdate()
	return connection.GetID()
}

func (cm *ServerManager) RemoveConnection(id string) {
	cm.connections[id].Disconnect()
	delete(cm.connections, id)
	cm.updateTrigger.SetPendingUpdate()
}

func (cm *ServerManager) GetConnections() []ServerInterface {
	connections := slices.Collect(maps.Values(cm.connections))
	slices.SortStableFunc(connections, func(a, b ServerInterface) int {
		if a.GetName() == b.GetName() {
			return strings.Compare(a.GetID(), b.GetID())
		}
		return strings.Compare(strings.ToLower(a.GetName()), strings.ToLower(b.GetName()))
	})
	return connections
}

func (cm *ServerManager) GetConnection(id string) ServerInterface {
	return cm.connections[id]
}

func (cm *ServerManager) Start() {
	for _, connection := range cm.connections {
		connection.Connect()
	}
}

func (cm *ServerManager) Stop() {
	for _, connection := range cm.connections {
		connection.Disconnect()
	}
}

func (cm *ServerManager) Load() {
	for _, server := range cm.config.Servers {
		if server.AutoConnect {
			cm.AddConnection(server.ID, server.Hostname, server.Port, server.TLS, server.Password, server.SASLLogin, server.SASLPassword, NewProfile(server.Profile.Nickname), false)
		}
	}
}

func (cm *ServerManager) SetUpdateTrigger(ut UpdateTrigger) {
	cm.updateTrigger = ut
}

func (cm *ServerManager) SetNotificationManager(nm NotificationManager) {
	cm.notificationManager = nm
	cm.commandManager.SetNotificationManager(nm)
}
