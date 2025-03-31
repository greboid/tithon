package main

import (
	"flag"
	"github.com/csmith/envflag"
	"github.com/greboid/ircclient/irc"
	"github.com/greboid/ircclient/web"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

var (
	debug     = flag.Bool("debug", true, "Show debugging")
	OpenUI    = flag.Bool("openui", true, "Should the UI launch")
	FixedPort = flag.Int("port", 8081, "Fixed port to use, 0 will use a random port")
)

func main() {
	envflag.Parse()
	options := &slog.HandlerOptions{}
	if *debug {
		options.Level = slog.LevelDebug
	} else {
		options.Level = slog.LevelInfo
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, options))
	slog.SetDefault(log)
	connectionManager := irc.NewConnectionManager()
	server := web.NewServer(connectionManager, irc.NewCommandManager(), *FixedPort)
	connectionManager.SetUpdateTrigger(server)
	defer server.Stop()
	defer connectionManager.Stop()
	err := connectionManager.Load()
	if err != nil {
		slog.Error("Unable to load config", "error", err)
		return
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGKILL, syscall.SIGINT)
	listenAddr := server.GetListenAddress()
	log.Info("Listening on", "address", listenAddr)
	go func() {
		connectionManager.Start()
	}()
	go func() {
		server.Start()
	}()
	if *OpenUI {

	}
	<-quit
	slog.Info("Quitting")
}
