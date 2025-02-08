package gui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"io/fs"
	"newirc/irc"
	"os"
)

type Config struct {
	Servers []irc.ConnectableServer `json:"servers"`
}

type App struct {
	Ctx         context.Context
	Connections []*irc.Client
}

func NewApp() *App {
	return &App{}
}

func (a *App) Shutdown(ctx context.Context) {
	for index := range a.Connections {
		a.Connections[index].Quit()
	}
	a.SaveConfig()
}

func (a *App) Startup(ctx context.Context) {
	a.Ctx = ctx
	a.LoadConfig()
}

func (a *App) LoadConfig() {
	data, err := os.ReadFile("config.json")
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		panic("Unable to load config: " + err.Error())
	}
	if data == nil {
		data = []byte("{}")
	}
	config := &Config{}
	err = json.Unmarshal(data, config)
	if err != nil {
		panic("Unable to load config: " + err.Error())
	}
	for i := range config.Servers {
		a.Connect(config.Servers[i])
	}
}

func (a *App) SaveConfig() {
	config := &Config{}
	for i := range a.Connections {
		runtime.LogDebugf(a.Ctx, "Adding server to config: %s", a.Connections[i].Server)
		config.Servers = append(config.Servers, irc.ConnectableServer{
			Server:       a.Connections[i].Server,
			TLS:          a.Connections[i].UseTLS,
			Saslusername: a.Connections[i].SASLLogin,
			Saslpassword: a.Connections[i].SASLPassword,
			Profile: irc.ConnectableProfile{
				Nick:     a.Connections[i].Nick,
				User:     a.Connections[i].User,
				Realname: a.Connections[i].RealName,
			},
		})
	}
	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		runtime.LogErrorf(a.Ctx, "Unable to marshall config: %s", err)
	}
	err = os.WriteFile("config.json", data, 0644)
	if err != nil {
		runtime.LogErrorf(a.Ctx, "Unable to save config: %s", err)
	}
}

func (a *App) Connect(server irc.ConnectableServer) (bool, error) {
	runtime.LogDebugf(a.Ctx, "Connecting to: %s", server.Server)
	client := &irc.Client{}
	client.Server = fmt.Sprintf("%s", server.Server)
	client.UseTLS = server.TLS
	client.SASLLogin = server.Saslusername
	client.SASLPassword = server.Saslpassword
	client.Nick = server.Profile.Nick
	client.Debug = true
	err := client.Connect()
	if err != nil {
		runtime.LogDebugf(a.Ctx, "Failed to connect: %s", err.Error())
		return false, err
	}
	runtime.LogDebugf(a.Ctx, "Connected to: %s", server.Server)
	go func() {
		client.Loop()
	}()
	a.Connections = append(a.Connections, client)
	runtime.EventsEmit(a.Ctx, "serverAdded", irc.Server{Name: client.Server})
	return true, nil
}

func (a *App) GetServers() []irc.Server {
	servers := make([]irc.Server, 0)
	for index := range a.Connections {
		servers = append(servers, irc.Server{Name: a.Connections[index].Server})
	}
	return servers
}

func (a *App) ExportTypesToWailsRuntime(ircmsg.Message) {}
