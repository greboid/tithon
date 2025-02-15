package irc

import (
	"context"
	"github.com/albinj12/unique-id"
	"github.com/ergochat/irc-go/ircevent"
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gopkg.in/yaml.v2"
	"newirc/events"
	"os"
	"time"
)

func (e NullEmitter) Emit(string, ...interface{}) {}

func NewApp() *App {
	return &App{}
}

func (a *App) Shutdown(context.Context) {
	a.EE = NullEmitter{}
}

func (a *App) Startup(ctx context.Context) {
	a.Ctx = ctx
	if a.Ctx.Value("events") != nil {
		a.EE = a.Ctx.Value("events").(EventEmitter)
	} else {
		a.EE = NullEmitter{}
	}
	a.LoadConfig()
}

func (a *App) UIReady() {
	for i := range a.Servers {
		a.EE.Emit("ServerAdded", events.ServerAdded{
			Server: events.Server{ID: a.Servers[i].ID(), Server: a.Servers[i].Name()},
			Time:   events.IRCTime{Time: time.Now()},
		})
	}
}

func (a *App) LoadConfig() {
	config := &Config{}
	data, err := os.ReadFile("./config.yaml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		panic(err)
	}
	for i := range config.Servers {
		_ = a.NewConnection(config.Servers[i].Server, config.Servers[i].TLS, config.Servers[i].SaslUsername, config.Servers[i].SaslPassword, config.Servers[i].Profile.Nick)
	}
}

func (a *App) SaveConfig() {
	config := &Config{}
	for range a.Servers {
		config.Servers = append(config.Servers, ConfigServer{
			Server:       "",
			TLS:          false,
			SaslUsername: "",
			SaslPassword: "",
			Profile: ConfigProfile{
				Nick: "",
				User: "",
			},
		})
	}
	data, err := yaml.Marshal(config)
	if err != nil {
		panic(err)
	}
	runtime.LogInfo(a.Ctx, string(data))
	//TODO: Save config
}

func (a *App) NewConnection(server string, useTLS bool, SASLLogin string, SASLPassword string, PreferredNick string) error {
	connection := &ConnectionImpl{}
	connection.Init(a.Ctx, server, useTLS, SASLLogin, SASLPassword, PreferredNick)
	a.Servers = append(a.Servers, connection)
	a.EE.Emit("ServerAdded", events.ServerAdded{
		Server: events.Server{ID: connection.ID(), Server: connection.Name()},
		Time:   events.IRCTime{Time: time.Now()},
	})
	a.SaveConfig()
	connection.Connect()
	connection.Loop()
	connection.connection.AddConnectCallback(func(message ircmsg.Message) {
		a.EE.Emit("ServerUpdated", events.ServerUpdated{
			Server: events.Server{ID: connection.ID(), Server: connection.Name()},
			Time:   events.IRCTime{Time: time.Now()},
		})
	})
	return nil
}

type ConnectionImpl struct {
	ee         EventEmitter
	connection *ircevent.Connection
	id         string
}

func (b *ConnectionImpl) Init(ctx context.Context, server string, useTLS bool, SASLLogin string, SASLPassword string, PreferredNick string) {
	if ctx.Value("events") != nil {
		b.ee = ctx.Value("events").(EventEmitter)
	} else {
		b.ee = NullEmitter{}
	}
	s, _ := uniqueid.Generateid()
	useSASL := SASLLogin != "" && SASLPassword != ""
	b.id = s
	b.connection = &ircevent.Connection{
		Server:   server,
		Nick:     PreferredNick,
		User:     "",
		RealName: "",
		Password: "",
		RequestCaps: []string{
			"message-tags",
			"echo-message",
			"server-time",
		},
		SASLLogin:    SASLLogin,
		SASLPassword: SASLPassword,
		QuitMessage:  " ",
		Version:      " ",
		UseTLS:       useTLS,
		UseSASL:      useSASL,
		//Log:          nil,
		Debug: true,
	}
}

func (b *ConnectionImpl) AddCallback(s string, f func(message ircmsg.Message)) {
	//TODO implement me
	panic("implement me")
}

func (b *ConnectionImpl) AddConnectCallback(f func(message ircmsg.Message)) {
	//TODO implement me
	panic("implement me")
}

func (b *ConnectionImpl) AddDisconnectCallback(f func(message ircmsg.Message)) {
	//TODO implement me
	panic("implement me")
}

func (b *ConnectionImpl) CurrentNick() string {
	return b.connection.CurrentNick()
}

func (b *ConnectionImpl) GetChanTypes() string {
	//TODO implement me
	panic("implement me")
}

func (b *ConnectionImpl) Connect() {
	go func() {
		err := b.connection.Connect()
		if err != nil {
			b.ee.Emit("ServerConnectionnError", events.ServerConnectionnError{
				Server: events.Server{Server: b.Name()},
				Time:   events.IRCTime{Time: time.Now()},
				Error:  err.Error(),
			})
		}
	}()
}

func (b *ConnectionImpl) Loop() {
	go func() {
		b.connection.Loop()
	}()
}

func (b *ConnectionImpl) Name() string {
	network := b.connection.ISupport()["NETWORK"]
	if network == "" {
		return b.connection.Server
	} else {
		return network
	}
}

func (b *ConnectionImpl) ID() string {
	return b.id
}
