package irc

import (
	"github.com/greboid/tithon/config"
	"log/slog"
	"maps"
	"regexp"
	"slices"
	"strings"
)

type UpdateTrigger interface {
	SetPendingUpdate()
}

type Notification struct {
	Text string
}

type NotificationManager struct {
	notifications        []config.NotificationTrigger
	pendingNotifications chan Notification
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

func NewNotificationManager(pendingNotifications chan Notification, triggers []config.NotificationTrigger) *NotificationManager {
	return &NotificationManager{
		pendingNotifications: pendingNotifications,
		notifications:        triggers,
	}
}

func (cm *NotificationManager) SendNotification(text string) {
	cm.pendingNotifications <- Notification{Text: text}
}

func (cm *NotificationManager) IsNotification(network, source, nick, message string) bool {
	for i := range cm.notifications {
		networkPattern := cm.notifications[i].Network
		if networkPattern == "" {
			networkPattern = ".*"
		}
		sourcePattern := cm.notifications[i].Source
		if sourcePattern == "" {
			sourcePattern = ".*"
		}
		nickPattern := cm.notifications[i].Nick
		if nickPattern == "" {
			nickPattern = ".*"
		}
		messagePattern := cm.notifications[i].Message
		if messagePattern == "" {
			messagePattern = ".*"
		}
		
		if regexp.MustCompile(networkPattern).MatchString(network) &&
			regexp.MustCompile(sourcePattern).MatchString(source) &&
			regexp.MustCompile(nickPattern).MatchString(nick) &&
			regexp.MustCompile(messagePattern).MatchString(message) {
			return true
		}
	}
	return false
}
