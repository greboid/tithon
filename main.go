package main

import (
	"flag"
	"github.com/csmith/envflag"
	"github.com/greboid/ircclient/irc"
	"github.com/greboid/ircclient/web"
	"github.com/pkg/browser"
	"github.com/webview/webview_go"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

//go:generate go run github.com/a-h/templ/cmd/templ@latest generate

var (
	debug       = flag.Bool("debug", false, "Show debugging")
	OpenBrowser = flag.Bool("openbrowser", false, "Should we open the browser")
	OpenUI      = flag.Bool("openui", true, "Should the UI launch")
	FixedPort   = flag.Int("port", 0, "Fixed port to use, 0 will use a random port")
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
	defer connectionManager.Stop()
	err := connectionManager.Load()
	if err != nil {
		slog.Error("Unable to load config", "error", err)
		return
	}
	server := web.NewServer(connectionManager, *FixedPort)
	defer server.Stop()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGKILL, syscall.SIGINT)
	listenAddr := server.GetListenAddress()
	go func() {
		connectionManager.Start()
	}()
	go func() {
		server.Start()
	}()
	if *OpenBrowser {
		go func() { _ = browser.OpenURL(listenAddr) }()
	}
	if *OpenUI {
		go func() {
			w := webview.New(false)
			defer w.Destroy()
			w.Dispatch(func() {
				w.SetTitle("IRC Client")
				w.SetSize(800, 600, webview.HintNone)
				w.Navigate(listenAddr)
			})
			w.Run()
			quit <- syscall.SIGINT
		}()
	}
	<-quit
	slog.Info("Quitting")
}
