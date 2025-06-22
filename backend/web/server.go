package web

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/greboid/tithon/config"
	"github.com/greboid/tithon/irc"
	semver "github.com/hashicorp/go-version"
	"html/template"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type SettingsData struct {
	Version         string
	TimestampFormat string
	ShowNicklist    bool
	Servers         []config.Server
	Notifications   []config.NotificationTrigger
	Theme           string
}

var (
	//go:embed static
	staticFS embed.FS
	//go:embed templates
	templateFS embed.FS
)

type WebClient struct {
	lock                 sync.Mutex
	httpServer           *http.Server
	connectionManager    *irc.ServerManager
	commands             *irc.CommandManager
	activeWindow         *irc.Window
	activeQuery          *irc.Query
	fixedPort            int
	templates            *template.Template
	activeLock           sync.Mutex
	serverList           *ServerList
	pendingUpdate        atomic.Bool
	windowChanged        atomic.Bool
	listlock             sync.Mutex
	uiUpdate             atomic.Bool
	pendingNotifications chan irc.Notification
	templateLock         sync.Mutex
	conf                 *config.Config
	inputHistory         []string
	historyPosition      int
	historyLock          sync.Mutex
	showSettings         chan bool
	settingsData         SettingsData
}

type ServerList struct {
	Parents     []*ServerListItem
	OrderedList []*ServerListItem
}

type ServerListItem struct {
	Window   *irc.Window
	Link     string
	Name     string
	Children []*ServerListItem
}

type inputValues struct {
	Input string `json:"input"`
}

func getVersion() string {
	var versionString string
	if info, ok := debug.ReadBuildInfo(); ok {
		versionString = info.Main.Version
		if version, err := semver.NewVersion(versionString); err == nil {
			versionString = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(version.Segments()[0:3])), "."), "[]")
			if version.Prerelease() != "" {
				versionString = versionString + "-dev"
			}
		} else {
			versionString = "err"
		}
	} else {
		versionString = "unknown"
	}
	return versionString
}

func NewWebClient(cm *irc.ServerManager, commands *irc.CommandManager, fixedPort int, pendingNotifications chan irc.Notification, conf *config.Config, showSettings chan bool) *WebClient {
	mux := http.NewServeMux()
	client := &WebClient{
		fixedPort: fixedPort,
		lock:      sync.Mutex{},
		httpServer: &http.Server{
			Handler: mux,
		},
		connectionManager:    cm,
		commands:             commands,
		activeWindow:         nil,
		serverList:           &ServerList{},
		pendingNotifications: pendingNotifications,
		conf:                 conf,
		inputHistory:         make([]string, 0),
		historyPosition:      -1,
		showSettings:         showSettings,
		settingsData: SettingsData{
			Version:         getVersion(),
			TimestampFormat: conf.UISettings.TimestampFormat,
			ShowNicklist:    conf.UISettings.ShowNicklist,
			Servers:         conf.Servers,
			Notifications:   conf.Notifications.Triggers,
		},
	}
	client.addRoutes(mux)
	return client
}

func (s *WebClient) GetListenAddress() (string, int) {
	if s.httpServer.Addr != "" {
		split := strings.Split(s.httpServer.Addr, ":")
		port, _ := strconv.Atoi(split[1])
		return split[0], port
	}
	ip, port, err := s.getPort()
	if err != nil {
		slog.Error("Unable to get free port", "error", err)
		return "", -1
	}
	s.httpServer.Addr = net.JoinHostPort(ip.String(), strconv.Itoa(port))
	return ip.String(), port
}

func (s *WebClient) Start() {
	if err := s.httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Error starting server:", slog.String("error", err.Error()))
	}
	slog.Debug("WebClient stopped")
}

func (s *WebClient) Stop() {
	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownRelease()
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("Error shutting down web server:", slog.String("error", err.Error()))
	}
}

func (s *WebClient) getPort() (net.IP, int, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("[::1]:%d", s.fixedPort))
	if err != nil {
		return nil, -1, err
	}
	listen, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, -1, err
	}
	defer func() { _ = listen.Close() }()
	lp := listen.Addr().(*net.TCPAddr)
	return lp.IP, lp.Port, nil
}

func (s *WebClient) getServerList() *ServerList {
	s.listlock.Lock()
	defer s.listlock.Unlock()
	s.serverList = &ServerList{}
	connections := s.connectionManager.GetConnections()
	for i := range connections {
		serverIndex := slices.IndexFunc(s.serverList.Parents, func(item *ServerListItem) bool {
			return item.Window == connections[i].Window
		})
		var server *ServerListItem
		if serverIndex == -1 {
			server = &ServerListItem{
				Window:   connections[i].Window,
				Link:     connections[i].GetID(),
				Name:     connections[i].GetName(),
				Children: nil,
			}
			s.serverList.Parents = append(s.serverList.Parents, server)
			s.serverList.OrderedList = append(s.serverList.OrderedList, server)
		} else {
			server = s.serverList.Parents[serverIndex]
		}
		channels := connections[i].GetChannels()
		for j := range channels {
			windowIndex := slices.IndexFunc(server.Children, func(item *ServerListItem) bool {
				return item.Window == channels[j].Window
			})
			if windowIndex == -1 {
				child := &ServerListItem{
					Window:   channels[j].Window,
					Link:     connections[i].GetID() + "/" + channels[j].GetID(),
					Name:     channels[j].GetName(),
					Children: nil,
				}
				server.Children = append(server.Children, child)
				s.serverList.OrderedList = append(s.serverList.OrderedList, child)
			}
		}

		queries := connections[i].GetQueries()
		for j := range queries {
			windowIndex := slices.IndexFunc(server.Children, func(item *ServerListItem) bool {
				return item.Window == queries[j].Window
			})
			if windowIndex == -1 {
				child := &ServerListItem{
					Window:   queries[j].Window,
					Link:     connections[i].GetID() + "/" + queries[j].GetID(),
					Name:     queries[j].GetName(),
					Children: nil,
				}
				server.Children = append(server.Children, child)
				s.serverList.OrderedList = append(s.serverList.OrderedList, child)
			}
		}
	}
	return s.serverList
}

func (s *WebClient) setActiveWindow(window *irc.Window) {
	s.activeLock.Lock()
	defer s.activeLock.Unlock()
	if s.activeWindow != nil {
		s.activeWindow.SetActive(false)
	}
	if window != nil {
		window.SetActive(true)
	}
	s.activeWindow = window
	s.SetPendingUpdate()
	s.SetWindowChanged()
}

func (s *WebClient) getActiveWindow() *irc.Window {
	s.activeLock.Lock()
	defer s.activeLock.Unlock()
	return s.activeWindow
}

func (s *WebClient) SetPendingUpdate() {
	s.pendingUpdate.Store(true)
}

func (s *WebClient) SetWindowChanged() {
	s.windowChanged.Store(true)
}

func (s *WebClient) SetUIUpdate() {
	s.uiUpdate.Store(true)
}
