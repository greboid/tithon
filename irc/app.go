package irc

import (
	"context"
	"fmt"
	"github.com/albinj12/unique-id"
	"github.com/ergochat/irc-go/ircevent"
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gopkg.in/yaml.v2"
	"newirc/events"
	"os"
	"slices"
	"strings"
	"time"
)

func (e NullEmitter) Emit(string, ...interface{})                                         {}
func (e NullEmitter) Add(eventName string, callback func(data ...interface{}))            {}
func (e NullEmitter) AddWithHistory(eventName string, callback func(data ...interface{})) {}

func NewApp() *App {
	return &App{}
}

func (a *App) Shutdown(context.Context) {
	a.EE = NullEmitter{}
}

func (a *App) Startup(ctx context.Context) {
	a.Ctx = ctx
	if a.Ctx.Value("events") != nil {
		a.EE = a.Ctx.Value("events").(WailsEmitter)
	} else {
		a.EE = NullEmitter{}
	}
	a.LoadConfig()
}

func (a *App) UIReady() {
	for i := range a.Servers {
		a.EE.Emit("ServerAdded", events.ServerAdded{
			Server: events.Server{ID: a.Servers[i].ID(), Server: a.Servers[i].Name(), Channels: make([]*events.Channel, 0)},
			Time:   events.IRCTime{Time: time.Now()},
		})
		channels := a.Servers[i].GetChannels()
		for j := range channels {
			a.EE.Emit("ChannelJoinedSelf", events.ChannelJoinedSelf{
				Channel: channels[j],
				Time:    events.IRCTime{Time: time.Now()},
			})
		}
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
		Server: events.Server{ID: connection.ID(), Server: connection.Name(), Channels: make([]*events.Channel, 0)},
		Time:   events.IRCTime{Time: time.Now()},
	})
	a.SaveConfig()
	connection.Connect()
	connection.Loop()
	connection.connection.AddConnectCallback(func(message ircmsg.Message) {
		a.EE.Emit("ServerUpdated", events.ServerUpdated{
			Server: events.Server{ID: connection.ID(), Server: connection.Name(), Channels: make([]*events.Channel, 0)},
			Time:   events.IRCTime{Time: time.Now()},
		})
	})
	return nil
}

type ConnectionImpl struct {
	ee         WailsEmitter
	connection *ircevent.Connection
	id         string
	channels   []events.Channel
}

func (b *ConnectionImpl) GetChannels() []events.Channel {
	return b.channels
}

func (b *ConnectionImpl) Init(ctx context.Context, server string, useTLS bool, SASLLogin string, SASLPassword string, PreferredNick string) {
	if ctx.Value("events") != nil {
		b.ee = ctx.Value("events").(WailsEmitter)
	} else {
		b.ee = NullEmitter{}
	}
	s, _ := uniqueid.Generateid("n")
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
	b.connection.AddCallback("JOIN", func(message ircmsg.Message) {
		if message.Nick() == b.CurrentNick() {
			b.SelfJoin(message.Params[0])
		}
	})
	b.connection.AddCallback("PART", func(message ircmsg.Message) {

	})
	b.connection.AddCallback("KICK", func(message ircmsg.Message) {

	})
	b.connection.AddCallback("QUIT", func(message ircmsg.Message) {

	})
	b.connection.AddCallback("PRIVMSG", func(message ircmsg.Message) {
		if strings.HasPrefix(message.Params[0], "#") {
			index := slices.IndexFunc(b.channels, func(channel events.Channel) bool {
				return channel.Name == message.Params[0]
			})
			nuh, err := message.NUH()
			if index != -1 && err == nil {

				b.ChannelMessage(events.ChannelMessage{
					Source: events.ChannelUser{
						User: events.User{
							Nick:     message.Nick(),
							UserHost: fmt.Sprintf("%s@%s", nuh.User, nuh.Host),
							Realname: "",
						},
						Modes: "",
					},
					Channel:  &b.channels[index],
					Message:  strings.Join(message.Params[1:], " "),
					IsNotice: false,
					IsAction: false,
				})
			}
		}
	})
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

func (b *ConnectionImpl) SelfJoin(name string) {
	channel := events.Channel{
		ServerID:           b.ID(),
		Name:               name,
		Users:              nil,
		Topic:              "",
		ModesList:          []events.ModeList{},
		ModesNoParam:       []events.ModeNoParam{},
		ModesParamSet:      []events.ModeParamSet{},
		ModesParamSetUnset: []events.ModeParamSetUnset{},
	}
	b.channels = append(b.channels, channel)
	b.ee.Emit("ChannelJoinedSelf", events.ChannelJoinedSelf{
		Channel: channel,
		Time:    events.IRCTime{Time: time.Now()},
	})
}

func (b *ConnectionImpl) ChannelMessage(message events.ChannelMessage) {
	b.ee.Emit("ChannelMessageReceived", events.ChannelMessageReceived{
		Message: message,
	})
}

func (a *App) ExportToWails(events.Server, events.Channel, events.ServerUpdated, events.ServerAdded, events.ChannelJoinedSelf, events.ChannelMessageReceived) {
}
