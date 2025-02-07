package main

import (
	"context"
	"embed"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"newirc/gui"
	"newirc/irc"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed all:build/appicon.png
var icon []byte

func main() {
	app := gui.NewApp()
	client := irc.Client{
		App: app,
	}

	err := wails.Run(&options.App{
		Title:     "IRC Client",
		MinWidth:  800,
		MinHeight: 600,
		Width:     1024,
		Height:    768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: func(ctx context.Context) {
			app.Ctx = ctx
			client.TestConnect()
		},
		OnShutdown: func(ctx context.Context) {
			client.Quit()
		},
		Bind: []interface{}{
			app,
		},
		Linux: &linux.Options{
			Icon:             icon,
			WebviewGpuPolicy: linux.WebviewGpuPolicyNever,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
