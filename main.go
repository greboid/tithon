package main

import (
	"flag"
	"github.com/csmith/envflag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"test/irc"
	"test/web"
)

//go:generate go run github.com/a-h/templ/cmd/templ@latest generate

var (
	debug = flag.Bool("debug", true, "Show debugging")
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
	err := connectionManager.Load()
	if err != nil {
		slog.Error("Unable to load config", "error", err)
		return
	}
	server := web.NewServer(connectionManager)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGKILL, syscall.SIGINT)
	go func() {
		connectionManager.Start()
	}()
	go func() {
		server.Start()
	}()
	<-quit
	connectionManager.Stop()
	server.Stop()
}
