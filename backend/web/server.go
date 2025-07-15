package web

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/greboid/tithon/config"
	"github.com/greboid/tithon/irc"
	"github.com/greboid/tithon/services"
	"html/template"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	//go:embed static
	staticFS embed.FS
	//go:embed templates
	templateFS embed.FS
)

type WebClient struct {
	lock                sync.Mutex
	httpServer          *http.Server
	connectionManager   *irc.ServerManager
	commands            *irc.CommandManager
	fixedPort           int
	templates           *template.Template
	pendingUpdate       atomic.Bool
	windowChanged       atomic.Bool
	uiUpdate            atomic.Bool
	templateLock        sync.Mutex
	conf                *config.Config
	showSettings        chan bool
	windowService       *services.WindowService
	serverListService   *services.ServerListService
	notificationService *services.NotificationService
	settingsService     *services.SettingsService
	inputHistoryService *services.InputHistoryService
}

type inputValues struct {
	Input string `json:"input"`
}

func NewWebClient(
	cm *irc.ServerManager,
	commands *irc.CommandManager,
	fixedPort int,
	conf *config.Config,
	showSettings chan bool,
	windowService *services.WindowService,
	serverListService *services.ServerListService,
	notificationService *services.NotificationService,
	settingsService *services.SettingsService,
	inputHistoryService *services.InputHistoryService,
) *WebClient {
	mux := http.NewServeMux()
	client := &WebClient{
		fixedPort: fixedPort,
		lock:      sync.Mutex{},
		httpServer: &http.Server{
			Handler: mux,
		},
		connectionManager:   cm,
		commands:            commands,
		conf:                conf,
		showSettings:        showSettings,
		windowService:       windowService,
		serverListService:   serverListService,
		notificationService: notificationService,
		settingsService:     settingsService,
		inputHistoryService: inputHistoryService,
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

func (s *WebClient) getServerList() *services.ServerList {
	return s.serverListService.GetServerList(s.connectionManager)
}

func (s *WebClient) setActiveWindow(window *irc.Window) {
	if s.windowService != nil {
		s.windowService.SetActiveWindow(window)
	}
}

func (s *WebClient) getActiveWindow() *irc.Window {
	if s.windowService != nil {
		return s.windowService.GetActiveWindow()
	}
	return nil
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

func (s *WebClient) OnWindowRemoved(removedWindow *irc.Window) {
	if s.windowService != nil {
		serverList := s.getServerList()
		s.windowService.OnWindowRemoved(removedWindow, serverList)
	}
}

func (s *WebClient) SetWindowService(windowService *services.WindowService) {
	s.windowService = windowService
}
