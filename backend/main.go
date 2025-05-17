package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/csmith/envflag/v2"
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
	OpenUI    = flag.Bool("openui", true, "Should the UI launch")
	FixedPort = flag.Int("port", 8081, "Fixed port to use, 0 will use a random port")
)

func main() {
	envflag.Parse()
	log := slogflags.Logger(
		slogflags.WithCustomLevels(map[string]slog.Level{"trace": irc.LevelTrace}),
		slogflags.WithSetDefault(true),
		slogflags.WithReplaceAttr(func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				switch {
				case level == irc.LevelTrace:
					a.Value = slog.StringValue("TRACE")
				}
			}
			return a
		}),
	)

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
