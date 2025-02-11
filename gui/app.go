package gui

import (
	"context"
	"errors"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gopkg.in/yaml.v3"
	"io/fs"
	"newirc/irc"
	"os"
)

type Config struct {
	Servers []irc.ConnectableServer `yaml:"servers" json:"servers"`
}

type App struct {
	Ctx         context.Context
	Connections []*irc.Client
}

func NewApp() *App {
	return &App{}
}

func (a *App) Shutdown(ctx context.Context) {
	runtime.LogInfof(ctx, "Shutting down")
	for index := range a.Connections {
		a.Connections[index].Quit()
	}
	a.SaveConfig()
}

func (a *App) Startup(ctx context.Context) {
	fmt.Println("Startup.")
	a.Ctx = ctx
}

func (a *App) Started() {
	a.LoadConfig()
}

func (a *App) LoadConfig() {
	data, err := os.ReadFile("config.yaml")
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		panic("Unable to load config: " + err.Error())
	}
	if data == nil {
		data = []byte("")
	}
	config := &Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		panic("Unable to load config: " + err.Error())
	}
	for i := range config.Servers {
		go func() {
			fmt.Println("Connecting")
			_, err = a.Connect(&config.Servers[i])
			if err != nil {
				runtime.LogDebugf(a.Ctx, "Failed to connect: %s", err.Error())
			}
		}()
	}
}

func (a *App) SaveConfig() {
}

func (a *App) Connect(server *irc.ConnectableServer) (bool, error) {
	runtime.LogDebugf(a.Ctx, "Connecting to: %s", server.Server)
	client := &irc.Client{
		Ctx:               a.Ctx,
		ConnectableServer: *server,
	}
	go runtime.EventsEmit(a.Ctx, "serverAdded", server)
	err := client.Connect(*server)
	if err != nil {
		runtime.LogDebugf(a.Ctx, "Failed to connect: %s", err.Error())
		return false, err
	}
	runtime.LogDebugf(a.Ctx, "Connected to: %s", server.Server)
	go func() {
		client.Loop()
	}()
	a.Connections = append(a.Connections, client)
	return true, nil
}

func (a *App) ExposeTypesToWails(profile irc.ConnectableProfile,
	server irc.ConnectableServer,
	channel irc.Channel,
	message irc.ChannelMessage,
	directMessage irc.DirectMessage,
	serverMessage irc.ServerMessage) {
}
