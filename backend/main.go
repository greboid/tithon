package main

import (
	"flag"
	"github.com/csmith/envflag"
	"github.com/csmith/slogflags"
	"github.com/greboid/tithon/config"
	"github.com/greboid/tithon/irc"
	"github.com/greboid/tithon/web"
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
	slogflags.Logger(slogflags.WithSetDefault(true))
	conf := &config.Config{}
	err := conf.Load()
	if err != nil {
		slog.Error("Unable to load config", "error", err)
		return
	}
	defer func() {
		err = conf.Save()
		if err != nil {
			slog.Error("Unable to save config", "error", err)
			return
		}
	}()
	connectionManager := irc.NewConnectionManager(conf)
	defer connectionManager.Stop()
	server := web.NewServer(connectionManager, irc.NewCommandManager(conf), *FixedPort)
	defer server.Stop()
	connectionManager.SetUpdateTrigger(server)
	connectionManager.Load()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGKILL, syscall.SIGINT)
	listenAddr := server.GetListenAddress()
	slog.Info("Listening on", "address", listenAddr)
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
