package web

import (
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"github.com/greboid/ircclient/irc"
	"github.com/pkg/browser"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	//go:embed static/*
	staticFS embed.FS

	fixedPort   = flag.Int("port", 0, "Fixed port to use")
	openBrowser = flag.Bool("openbrowser", false, "Should we open the browser")
)

type Server struct {
	lock              sync.Mutex
	httpServer        *http.Server
	connectionManager *irc.ConnectionManager
	activeServer      string
	activeWindow      string
}

func NewServer(cm *irc.ConnectionManager) *Server {
	mux := http.NewServeMux()
	server := &Server{
		lock: sync.Mutex{},
		httpServer: &http.Server{
			Handler: mux,
		},
		connectionManager: cm,
	}
	server.addRoutes(mux)
	return server
}

func (s *Server) GetListenAddress() string {
	if s.httpServer.Addr != "" {
		return fmt.Sprintf("http://%s", s.httpServer.Addr)
	}
	listenAddr, clickAddr, err := getListenAddr()
	if err != nil {
		slog.Error("Unable to get free port", "error", err)
		return ""
	}
	s.httpServer.Addr = listenAddr
	return clickAddr
}

func (s *Server) Start() string {
	clickAddr := s.GetListenAddress()
	slog.Info("Starting webserver", "url", clickAddr)
	if *openBrowser {
		go func() { _ = browser.OpenURL(clickAddr) }()
	}
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

func getListenAddr() (string, string, error) {
	ip, port, err := getPort()
	if err != nil {
		return "", "", err
	}
	if ip.To4() == nil {
		return fmt.Sprintf("[%s]:%d", ip, port), fmt.Sprintf("http://[%s]:%d", ip, port), nil
	}
	return fmt.Sprintf("%s:%d", ip, port), fmt.Sprintf("http://%s:%d", ip, port), nil
}

func getPort() (net.IP, int, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("[::1]:%d", *fixedPort))
	if err != nil {
		addr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("127.0.0.1:%d", *fixedPort))
		if err != nil {
			return nil, -1, err
		}
	}
	listen, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, -1, err
	}
	defer func() { _ = listen.Close() }()
	lp := listen.Addr().(*net.TCPAddr)
	return lp.IP, lp.Port, nil
}
