package main

import (
	"flag"
	"github.com/csmith/envflag/v2"
	"github.com/csmith/slogflags"
	"github.com/greboid/tithon/config"
	"github.com/greboid/tithon/irc"
	"github.com/greboid/tithon/services"
	"github.com/greboid/tithon/web"
	"github.com/hueristiq/hq-go-url/extractor"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

var (
	FixedPort = flag.Int("port", 0, "Fixed port to use, 0 will use a random port")
)

func main() {
	envflag.Parse()
	slogflags.Logger(
		slogflags.WithCustomLevels(map[string]slog.Level{"trace": irc.LevelTrace}),
		slogflags.WithSetDefault(true),
	)

	provider, err := config.NewDefaultConfigProvider()
	if err != nil {
		slog.Error("Unable to load config", "error", err)
		return
	}
	conf := config.NewConfig(provider)
	err = conf.Load()
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
	linkRegex := extractor.New(extractor.WithHost()).CompileRegex()
	showSettings := make(chan bool)
	pendingNotifications := make(chan irc.Notification, 10000)
	notificationManager := irc.NewNotificationManager(pendingNotifications, conf.Notifications.Triggers)
	commandManager := irc.NewCommandManager(linkRegex, conf, showSettings)
	connectionManager := irc.NewServerManager(linkRegex, conf, commandManager)
	defer connectionManager.Stop()

	settingsService := services.NewSettingsService(conf)
	inputHistoryService := services.NewInputHistoryService()
	serverListService := services.NewServerListService()
	notificationService := services.NewNotificationService(pendingNotifications)

	server := web.NewWebClient(
		connectionManager,
		commandManager,
		*FixedPort,
		conf,
		showSettings,
		nil,
		serverListService,
		notificationService,
		settingsService,
		inputHistoryService,
	)
	defer server.Stop()

	windowService := services.NewWindowService(server, server)
	server.SetWindowService(windowService)
	connectionManager.SetUpdateTrigger(server)
	connectionManager.SetNotificationManager(notificationManager)
	connectionManager.SetWindowRemovalCallback(server)
	connectionManager.Load()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGKILL, syscall.SIGINT)
	host, port := server.GetListenAddress()
	slog.Info("Listening on", "host", host, "port", port)
	go func() {
		connectionManager.Start()
	}()
	go func() {
		server.Start()
	}()
	<-quit
	slog.Info("Quitting")
}
