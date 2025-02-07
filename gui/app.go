package gui

import (
	"context"
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"newirc/irc"
)

type App struct {
	Ctx         context.Context
	Connections []irc.Client
}

func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.Ctx = ctx
}

func (a *App) applicationMenu() *menu.Menu {
	AppMenu := menu.NewMenu()
	FileMenu := AppMenu.AddSubmenu("File")
	FileMenu.AddText("Quit", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		runtime.Quit(a.Ctx)
	})
	return AppMenu
}

func (a *App) Connect(server irc.Server, profile irc.Profile) {}

func (a *App) ExportTypesToWailsRuntime(ircmsg.Message) {}
