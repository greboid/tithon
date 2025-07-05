package irc

import (
	"github.com/greboid/tithon/config"
	"github.com/hueristiq/hq-go-url/extractor"
	"maps"
	"regexp"
	"slices"
	"strings"
)

var linkRegex = extractor.New(extractor.WithHost()).CompileRegex()

type UpdateTrigger interface {
	SetPendingUpdate()
}

type WindowRemovalCallback interface {
	OnWindowRemoved(window *Window)
}

type ServerManager struct {
	connections           map[string]*Server
	commandManager        *CommandManager
	updateTrigger         UpdateTrigger
	notificationManager   NotificationManager
	timestampFormat       string
	linkRegex             *regexp.Regexp
	windowRemovalCallback WindowRemovalCallback
}

func NewServerManager(timestampFormat string, commandManager *CommandManager) *ServerManager {
	return &ServerManager{
		connections:     map[string]*Server{},
		commandManager:  commandManager,
		timestampFormat: timestampFormat,
		linkRegex:       linkRegex,
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
	connection := NewServer(cm.timestampFormat, id, hostname, port, tls, password, sasllogin, saslpassword, profile, cm.updateTrigger, cm.notificationManager)
	if cm.windowRemovalCallback != nil {
		connection.SetWindowRemovalCallback(cm.windowRemovalCallback)
	}
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
	connection := cm.connections[id]
	if connection != nil && cm.windowRemovalCallback != nil {
		cm.windowRemovalCallback.OnWindowRemoved(connection.GetWindow())
	}
	connection.Disconnect()
	delete(cm.connections, id)
	cm.updateTrigger.SetPendingUpdate()
}

func (cm *ServerManager) GetConnections() []*Server {
	connections := slices.Collect(maps.Values(cm.connections))
	slices.SortStableFunc(connections, func(a, b *Server) int {
		if a.GetName() == b.GetName() {
			return strings.Compare(a.GetID(), b.GetID())
		}
		return strings.Compare(strings.ToLower(a.GetName()), strings.ToLower(b.GetName()))
	})
	return connections
}

func (cm *ServerManager) GetConnection(id string) *Server {
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

func (cm *ServerManager) Load(servers []config.Server) {
	for _, server := range servers {
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

func (cm *ServerManager) SetWindowRemovalCallback(callback WindowRemovalCallback) {
	cm.windowRemovalCallback = callback
}
