package web

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/greboid/ircclient/irc"
	"html/template"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var (
	//go:embed static
	staticFS embed.FS
	//go:embed templates
	templateFS embed.FS
)

type Server struct {
	lock                 sync.Mutex
	httpServer           *http.Server
	connectionManager    *irc.ConnectionManager
	commands             *irc.CommandManager
	activeWindow         *irc.Window
	activePrivateMessage *irc.PrivateMessage
	fixedPort            int
	templates            *template.Template
	activeLock           sync.Mutex
}

func NewServer(cm *irc.ConnectionManager, commands *irc.CommandManager, fixedPort int) *Server {
	mux := http.NewServeMux()
	server := &Server{
		fixedPort: fixedPort,
		lock:      sync.Mutex{},
		httpServer: &http.Server{
			Handler: mux,
		},
		connectionManager: cm,
		commands:          commands,
	}
	server.addRoutes(mux)
	return server
}

func (s *Server) GetListenAddress() string {
	if s.httpServer.Addr != "" {
		return fmt.Sprintf("http://%s", s.httpServer.Addr)
	}
	ip, port, err := s.getPort()
	if err != nil {
		slog.Error("Unable to get free port", "error", err)
		return ""
	}
	s.httpServer.Addr = net.JoinHostPort(ip.String(), strconv.Itoa(port))
	return fmt.Sprintf("http://%s", s.httpServer.Addr)
}

func (s *Server) Start() string {
	clickAddr := s.GetListenAddress()
	slog.Info("Starting webserver", "url", clickAddr)
	if err := s.httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Error starting server:", slog.String("error", err.Error()))
	}
	slog.Debug("Server stopped")
	return clickAddr
}

func (s *Server) Stop() {
	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownRelease()
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("Error shutting down server:", slog.String("error", err.Error()))
	}
}

func (s *Server) getPort() (net.IP, int, error) {
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

func (s *Server) setActiveWindow(window *irc.Window) {
	s.activeLock.Lock()
	defer s.activeLock.Unlock()
	if s.activeWindow != nil {
		s.activeWindow.SetActive(false)
	}
	if window != nil {
		window.SetActive(true)
		window.SetUnread(false)
	}
	s.activeWindow = window
}

func (s *Server) getActiveWindow() *irc.Window {
	s.activeLock.Lock()
	defer s.activeLock.Unlock()
	return s.activeWindow
}
