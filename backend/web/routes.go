package web

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	uniqueid "github.com/albinj12/unique-id"
	"github.com/fsnotify/fsnotify"
	"github.com/greboid/tithon/config"
	"github.com/greboid/tithon/irc"
	"github.com/kirsle/configdir"
	datastar "github.com/starfederation/datastar/sdk/go"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

func getTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"map": func(pairs ...any) (map[string]interface{}, error) {
			if len(pairs)%2 != 0 {
				return nil, errors.New("incorrect number of arguments")
			}

			m := make(map[string]interface{}, len(pairs)/2)
			for i := 0; i < len(pairs); i += 2 {
				k, ok := pairs[i].(string)
				if !ok {
					return nil, errors.New("map keys must be strings")
				}
				m[k] = pairs[i+1]
			}

			return m, nil
		},
		"arr": func(elements ...any) []interface{} {
			return elements
		},
		"unsafe": func(input string) template.HTML {
			return template.HTML(input)
		},
	}
}

func (s *WebClient) noCacheMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		h.ServeHTTP(w, r)
	})
}

func (s *WebClient) addRoutes(mux *http.ServeMux) {
	var static fs.FS
	if stat, err := os.Stat("./web/static"); err == nil && stat.IsDir() {
		slog.Debug("Using on disk static resources")
		static = os.DirFS("./web/static")
		s.createStaticWatcher()
	} else {
		slog.Debug("Using on embedded static resources")
		static, _ = fs.Sub(staticFS, "static")
	}
	usercss := filepath.Join(configdir.LocalConfig("tithon"), "user.css")
	if _, err := os.Stat(usercss); err != nil {
		if _, err = os.OpenFile(usercss, os.O_CREATE, 0600); err != nil {
			slog.Debug("Unable to create empty user.css")
		}
	}
	s.createUserCSSWatcher(usercss)
	var allTemplates fs.FS
	if stat, err := os.Stat("./web/templates"); err == nil && stat.IsDir() {
		slog.Debug("Using on disk templates")
		allTemplates = os.DirFS("./web/templates")
		s.createTemplateWatcher(allTemplates)
	} else {
		slog.Debug("Using on embedded templates")
		allTemplates, _ = fs.Sub(templateFS, "templates")
	}
	s.updateTemplates(allTemplates)
	mux.Handle("GET /static/", s.noCacheMiddleware(http.StripPrefix("/static/", http.FileServer(http.FS(static)))))
	mux.HandleFunc("GET /static/user.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, usercss)
	})
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("GET /update", s.handleUpdate)
	mux.HandleFunc("GET /showSettings", s.handleShowSettings)
	mux.HandleFunc("GET /saveSettings", s.handleSaveSettings)
	mux.HandleFunc("GET /showAddServer", s.handleShowAddServer)
	mux.HandleFunc("GET /deleteServer", s.handleDeleteServer)
	mux.HandleFunc("GET /addServer", s.handleAddServer)
	mux.HandleFunc("GET /showEditServer", s.handleShowEditServer)
	mux.HandleFunc("GET /editServer", s.handleEditServer)
	mux.HandleFunc("GET /connectServer", s.handleConnectServer)
	mux.HandleFunc("GET /cancelEditServer", s.handleDefaultSettings)
	mux.HandleFunc("GET /changeWindow/{server}", s.handleChangeServer)
	mux.HandleFunc("GET /changeWindow/{server}/{channel}", s.handleChangeChannel)
	mux.HandleFunc("GET /s/{server}", s.handleServer)
	mux.HandleFunc("GET /s/{server}/{channel}", s.handleChannel)
	mux.HandleFunc("GET /input", s.handleInput)
	mux.HandleFunc("POST /upload", s.handleUpload)
	mux.HandleFunc("GET /join", s.handleJoin)
	mux.HandleFunc("GET /part", s.handlePart)
	mux.HandleFunc("GET /nextWindowUp", s.handleNextWindowUp)
	mux.HandleFunc("GET /nextWindowDown", s.handleNextWindowDown)
	mux.HandleFunc("GET /tab", s.handleTab)
	mux.HandleFunc("GET /nicklistshow", s.handleUpdateNicklist)
	mux.HandleFunc("GET /historyUp", s.handleHistoryUp)
	mux.HandleFunc("GET /historyDown", s.handleHistoryDown)
}

func (s *WebClient) createTemplateWatcher(templates fs.FS) {
	templateWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("Unable to create watcher", "error", err)
	}
	go s.templateLoop(templateWatcher, templates)
	err = templateWatcher.Add("./web/templates")
	if err != nil {
		slog.Error("Error add template watcher", "error", err)
	}
}

func (s *WebClient) createStaticWatcher() {
	staticWatches, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("Unable to create watcher", "error", err)
	}
	go s.staticLoop(staticWatches)
	err = staticWatches.Add("./web/static")
	if err != nil {
		slog.Error("Error add static watcher", "error", err)
	}
}

func (s *WebClient) createUserCSSWatcher(usercss string) {
	staticWatches, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("Unable to create watcher", "error", err)
	}
	go s.staticLoop(staticWatches)
	if _, err = os.UserConfigDir(); err == nil {
		err = staticWatches.Add(usercss)
		if err != nil {
			slog.Error("Error add static watcher", "error", err)
		}
	}
}

func (s *WebClient) staticLoop(watcher *fsnotify.Watcher) {
	defer func() {
		_ = watcher.Close()
	}()
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) {
				s.SetUIUpdate()
				slog.Debug("Updating static files", "file", event.Name)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			slog.Error("error listening for static changes:", "error", err)
		}
	}
}

func (s *WebClient) templateLoop(watcher *fsnotify.Watcher, templates fs.FS) {
	defer func() {
		_ = watcher.Close()
	}()
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) {
				s.updateTemplates(templates)
				slog.Debug("Updating templates", "file", event.Name)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			slog.Error("error listening for template changes:", "error", err)
		}
	}
}

func (s *WebClient) updateTemplates(allTemplates fs.FS) {
	allParsedTemplates, err := template.New("").Funcs(getTemplateFuncs()).ParseFS(allTemplates, "*.gohtml")
	if err != nil {
		slog.Error("Error parsing templates", "error", err)
		panic("Unable to load templates")
	}
	s.templateLock.Lock()
	defer s.templateLock.Unlock()
	s.templates = allParsedTemplates
	s.SetPendingUpdate()
}

func (s *WebClient) handleIndex(w http.ResponseWriter, _ *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pendingUpdate.Store(true)
	err := s.templates.ExecuteTemplate(w, "Index.gohtml", getVersion())
	if err != nil {
		slog.Debug("Error serving index", "error", err)
		return
	}
}

func (s *WebClient) updateMainUI(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()
	sse := datastar.NewSSE(w, r)
	err := sse.ExecuteScript("window.location.reload()", datastar.WithExecuteScriptAutoRemove(true))
	if err != nil {
		slog.Debug("Error refreshing page", "error", err)
		return
	}
}

func (s *WebClient) UpdateUI(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()
	sse := datastar.NewSSE(w, r)
	dataBytes, _ := json.Marshal(struct {
		Show bool `json:"nicklistshow"`
	}{s.conf.UISettings.ShowNicklist})
	var data bytes.Buffer
	data.WriteString(datastar.DoubleNewLine)
	err := sse.MergeSignals(dataBytes, datastar.WithOnlyIfMissing(true))
	if err != nil {
		slog.Debug("Error merging nicklist signal", "error", err)
	}
	s.outputTemplate(&data, "Serverlist.gohtml", s.getServerList())
	s.outputTemplate(&data, "Nicksettings.gohtml", nil)
	if s.getActiveWindow() == nil {
		s.outputTemplate(&data, "WindowInfo.gohtml", "")
		s.outputTemplate(&data, "Messages.gohtml", nil)
		s.outputTemplate(&data, "Nicklist.gohtml", nil)
	} else {
		s.outputTemplate(&data, "WindowInfo.gohtml", s.getActiveWindow().GetTitle())
		s.outputTemplate(&data, "Messages.gohtml", s.getActiveWindow().GetMessages())
		s.outputTemplate(&data, "Nicklist.gohtml", s.getActiveWindow().GetUsers())
	}

	err = sse.MergeFragments(data.String())
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
	err = sse.ExecuteScript(`document.getElementById("messages").scrollTo(0, document.getElementById("messages").scrollHeight)`, datastar.WithExecuteScriptAutoRemove(true))
	if err != nil {
		slog.Debug("Error scrolling to bottom", "error", err)
		return
	}
	var fileHost string
	if s.getActiveWindow() == nil || s.getActiveWindow().GetServer() == nil {
		fileHost = ""
	} else {
		fileHost = s.getActiveWindow().GetServer().GetFileHost()
	}
	type FileHost struct {
		Url string `json:"filehost"`
	}
	jsonData, _ := json.Marshal(FileHost{Url: fileHost})
	err = sse.MergeSignals(jsonData)
	if err != nil {
		slog.Debug("Error merging signals", "error", err)
		return
	}
}

func (s *WebClient) outputTemplate(wr io.Writer, name string, data any) {
	s.templateLock.Lock()
	defer s.templateLock.Unlock()
	err := s.templates.ExecuteTemplate(wr, name, data)
	if err != nil {
		slog.Debug("Error generating name", "error", err)
		return
	}
}

func (s *WebClient) handleUpdate(w http.ResponseWriter, r *http.Request) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			slog.Debug("Client connection closed")
			return
		case <-ticker.C:
			if yes := s.uiUpdate.Swap(false); yes {
				s.pendingUpdate.Store(false)
				s.updateMainUI(w, r)
			}
			if yes := s.pendingUpdate.Swap(false); yes {
				s.UpdateUI(w, r)
			}
		case <-s.showSettings:
			s.handleShowSettings(w, r)
		case notification := <-s.pendingNotifications:
			slog.Debug("Sending notification", "notification", notification)
			err := datastar.NewSSE(w, r).ExecuteScript(fmt.Sprintf(`notify("%s", "%s", %t, %t, "")`, notification.Title, notification.Text, notification.Popup, notification.Sound), datastar.WithExecuteScriptAutoRemove(true))
			if err != nil {
				slog.Error("Unable to send notification", "error", err)
			}
		}
	}
}

func (s *WebClient) handleShowSettings(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()
	sse := datastar.NewSSE(w, r)
	slog.Debug("Showing settings")
	var data bytes.Buffer

	s.settingsData = SettingsData{
		Version:         getVersion(),
		TimestampFormat: s.conf.UISettings.TimestampFormat,
		Theme:           s.conf.UISettings.Theme,
		ShowNicklist:    s.conf.UISettings.ShowNicklist,
		Notifications:   s.conf.Notifications.Triggers,
	}
	for i := range s.conf.Servers {
		s.settingsData.Servers = append(s.settingsData.Servers, config.Server{
			ID:           s.conf.Servers[i].ID,
			Hostname:     s.conf.Servers[i].Hostname,
			Port:         s.conf.Servers[i].Port,
			TLS:          s.conf.Servers[i].TLS,
			Password:     s.conf.Servers[i].Password,
			SASLLogin:    s.conf.Servers[i].SASLLogin,
			SASLPassword: s.conf.Servers[i].SASLPassword,
			Profile: config.Profile{
				Nickname: s.conf.Servers[i].Profile.Nickname,
			},
			AutoConnect: s.conf.Servers[i].AutoConnect,
		})
	}

	err := s.templates.ExecuteTemplate(&data, "SettingsPage.gohtml", nil)
	if err != nil {
		slog.Debug("Error generating template", "error", err)
	}
	err = s.templates.ExecuteTemplate(&data, "SettingsContent.gohtml", s.settingsData)
	if err != nil {
		slog.Debug("Error generating template", "error", err)
	}
	err = sse.MergeFragments(data.String())
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
}

func (s *WebClient) handleDefaultSettings(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()
	sse := datastar.NewSSE(w, r)
	slog.Debug("Showing settings")
	var data bytes.Buffer

	err := s.templates.ExecuteTemplate(&data, "SettingsContent.gohtml", s.settingsData)
	if err != nil {
		slog.Debug("Error generating template", "error", err)
	}
	err = sse.MergeFragments(data.String())
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
}

func (s *WebClient) handleDeleteServer(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()
	sse := datastar.NewSSE(w, r)
	slog.Debug("Showing settings")
	var data bytes.Buffer
	serverID := r.URL.Query().Get("id")

	s.settingsData.Servers = slices.DeleteFunc(s.settingsData.Servers, func(server config.Server) bool {
		return server.ID == serverID
	})

	err := s.templates.ExecuteTemplate(&data, "SettingsContent.gohtml", s.settingsData)
	if err != nil {
		slog.Debug("Error generating template", "error", err)
	}
	err = sse.MergeFragments(data.String())
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
}

func (s *WebClient) handleConnectServer(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()
	sse := datastar.NewSSE(w, r)
	slog.Debug("Connecting to server")

	serverID := r.URL.Query().Get("id")

	index := slices.IndexFunc(s.settingsData.Servers, func(server config.Server) bool {
		return server.ID == serverID
	})

	if index == -1 {
		slog.Debug("Unknown server specified", "WebClient ID", serverID)
		return
	}

	s.connectionManager.AddConnection(
		s.settingsData.Servers[index].ID,
		s.settingsData.Servers[index].Hostname,
		s.settingsData.Servers[index].Port,
		s.settingsData.Servers[index].TLS,
		s.settingsData.Servers[index].Password,
		s.settingsData.Servers[index].SASLLogin,
		s.settingsData.Servers[index].SASLPassword,
		irc.NewProfile(s.settingsData.Servers[index].Profile.Nickname),
		true,
	)

	var data bytes.Buffer
	err := s.templates.ExecuteTemplate(&data, "SettingsContent.gohtml", s.settingsData)
	if err != nil {
		slog.Debug("Error generating template", "error", err)
	}
	err = sse.MergeFragments(data.String())
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
}

func (s *WebClient) handleShowAddServer(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()
	sse := datastar.NewSSE(w, r)
	slog.Debug("Showing settings")
	var data bytes.Buffer
	err := s.templates.ExecuteTemplate(&data, "AddServerPage.gohtml", nil)
	if err != nil {
		slog.Debug("Error generating template", "error", err)
	}
	err = sse.MergeFragments(data.String())
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
}

func (s *WebClient) handleAddServer(w http.ResponseWriter, r *http.Request) {
	hostname := r.URL.Query().Get("hostname")
	port := r.URL.Query().Get("port")
	var err error
	portInt := 6697
	portInt, _ = strconv.Atoi(port)
	tlsBool := true
	tls := r.URL.Query().Get("tls")
	if tls == "" {
		tlsBool = false
	}
	tlsBool, _ = strconv.ParseBool(tls)
	nickname := r.URL.Query().Get("nickname")
	sasllogin := r.URL.Query().Get("sasllogin")
	saslpassword := r.URL.Query().Get("saslpassword")
	password := r.URL.Query().Get("password")
	autoConnectBool := true
	autoConnect := r.URL.Query().Get("connect")
	if autoConnect == "" {
		autoConnectBool = false
	}
	id, _ := uniqueid.Generateid("a", 5, "s")
	s.settingsData.Servers = append(s.settingsData.Servers, config.Server{
		Hostname:     hostname,
		Port:         portInt,
		TLS:          tlsBool,
		Password:     password,
		SASLLogin:    sasllogin,
		SASLPassword: saslpassword,
		Profile: config.Profile{
			Nickname: nickname,
		},
		AutoConnect: autoConnectBool,
		ID:          id,
	})

	s.lock.Lock()
	defer s.lock.Unlock()
	sse := datastar.NewSSE(w, r)
	var data bytes.Buffer
	err = s.templates.ExecuteTemplate(&data, "SettingsContent.gohtml", s.settingsData)
	if err != nil {
		slog.Debug("Error generating template", "error", err)
	}
	err = sse.MergeFragments(data.String())
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
}

func (s *WebClient) handleShowEditServer(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()
	sse := datastar.NewSSE(w, r)
	slog.Debug("Showing edit server dialog")

	serverID := r.URL.Query().Get("id")

	index := slices.IndexFunc(s.settingsData.Servers, func(server config.Server) bool {
		return server.ID == serverID
	})
	con := s.settingsData.Servers[index]
	if index == -1 {
		slog.Debug("Unknown server specified", "WebClient ID", serverID)
		return
	}

	serverToEdit := config.Server{
		Hostname:     con.Hostname,
		Port:         con.Port,
		TLS:          con.TLS,
		Password:     con.Password,
		SASLLogin:    con.SASLLogin,
		SASLPassword: con.SASLPassword,
		Profile: config.Profile{
			Nickname: con.Profile.Nickname,
		},
		ID:          con.ID,
		AutoConnect: con.AutoConnect,
	}

	var data bytes.Buffer
	err := s.templates.ExecuteTemplate(&data, "EditServerPage.gohtml", serverToEdit)
	if err != nil {
		slog.Debug("Error generating template", "error", err)
	}
	err = sse.MergeFragments(data.String())
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
}

func (s *WebClient) handleEditServer(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	hostname := r.URL.Query().Get("hostname")
	port := r.URL.Query().Get("port")
	var err error
	portInt := 6697
	portInt, _ = strconv.Atoi(port)
	tlsBool := true
	tls := r.URL.Query().Get("tls")
	if tls == "" {
		tlsBool = false
	}
	tlsBool, _ = strconv.ParseBool(tls)
	nickname := r.URL.Query().Get("nickname")
	sasllogin := r.URL.Query().Get("sasllogin")
	saslpassword := r.URL.Query().Get("saslpassword")
	password := r.URL.Query().Get("password")
	autoConnectBool := true
	autoConnect := r.URL.Query().Get("connect")
	if autoConnect == "" {
		autoConnectBool = false
	}

	for i := range s.settingsData.Servers {
		if s.settingsData.Servers[i].ID == id {
			s.settingsData.Servers[i].Hostname = hostname
			s.settingsData.Servers[i].Port = portInt
			s.settingsData.Servers[i].TLS = tlsBool
			s.settingsData.Servers[i].Password = password
			s.settingsData.Servers[i].SASLLogin = sasllogin
			s.settingsData.Servers[i].SASLPassword = saslpassword
			s.settingsData.Servers[i].Profile.Nickname = nickname
			s.settingsData.Servers[i].AutoConnect = autoConnectBool
		}
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	sse := datastar.NewSSE(w, r)
	var data bytes.Buffer
	err = s.templates.ExecuteTemplate(&data, "SettingsContent.gohtml", s.settingsData)
	if err != nil {
		slog.Debug("Error generating template", "error", err)
	}
	err = sse.MergeFragments(data.String())
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
}

func (s *WebClient) handleSaveSettings(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()
	slog.Debug("Saving settings")
	timestampFormat := r.URL.Query().Get("timestampFormat")
	showNicklist := r.URL.Query().Get("showNicklist") == "on"
	theme := r.URL.Query().Get("theme")
	if theme == "" {
		theme = "auto"
	}

	s.conf.UISettings.TimestampFormat = timestampFormat
	s.conf.UISettings.ShowNicklist = showNicklist
	s.conf.UISettings.Theme = theme
	s.conf.Notifications.Triggers = s.settingsData.Notifications
	s.conf.Servers = []config.Server{}
	for i := range s.settingsData.Servers {
		s.conf.Servers = append(s.conf.Servers, config.Server{
			Hostname:     s.settingsData.Servers[i].Hostname,
			Port:         s.settingsData.Servers[i].Port,
			TLS:          s.settingsData.Servers[i].TLS,
			Password:     s.settingsData.Servers[i].Password,
			SASLLogin:    s.settingsData.Servers[i].SASLLogin,
			SASLPassword: s.settingsData.Servers[i].SASLPassword,
			Profile: config.Profile{
				Nickname: s.settingsData.Servers[i].Profile.Nickname,
			},
			ID:          s.settingsData.Servers[i].ID,
			AutoConnect: s.settingsData.Servers[i].AutoConnect,
		})
	}

	err := s.conf.Save()
	if err != nil {
		slog.Error("Error saving config", "error", err)
	}

	sse := datastar.NewSSE(w, r)
	var data bytes.Buffer
	err = s.templates.ExecuteTemplate(&data, "EmptyDialog.gohtml", nil)
	if err != nil {
		slog.Debug("Error generating template", "error", err)
	}
	err = sse.ExecuteScript(`document.getElementById('dialog').close()`)
	if err != nil {
		slog.Debug("Error executing script", "error", err)
		return
	}
	err = sse.MergeFragments(data.String())
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
}

func (s *WebClient) handleChannel(w http.ResponseWriter, r *http.Request) {
	serverID := r.PathValue("server")
	channelName := r.PathValue("channel")
	connection := s.connectionManager.GetConnection(serverID)
	if connection == nil {
		slog.Debug("Invalid change channel call, unknown server", "server", serverID)
		s.setActiveWindow(nil)
		s.handleIndex(w, r)
		return
	}
	channel, _ := connection.GetChannelByName(channelName)
	if channel == nil {
		slog.Debug("Invalid change channel call, unknown channel", "server", serverID, "channel", channelName)
		s.setActiveWindow(nil)
		s.handleIndex(w, r)
		return
	}
	s.setActiveWindow(channel.Window)
	slog.Debug("Changing Window", "window", channel.Window.GetID())
	s.handleIndex(w, r)
}

func (s *WebClient) handleChangeChannel(w http.ResponseWriter, r *http.Request) {
	serverID := r.PathValue("server")
	channelID := r.PathValue("channel")
	connection := s.connectionManager.GetConnection(serverID)
	if connection == nil {
		slog.Debug("Invalid change channel call, unknown server", "server", serverID)
		return
	}
	channel := connection.GetChannel(channelID)
	if channel == nil {
		privateMessage := connection.GetQuery(channelID)
		if privateMessage == nil {
			slog.Debug("Invalid change channel call, unknown channel or private message", "server", serverID, "id", channelID)
			return
		}
		s.setActiveWindow(privateMessage.Window)
		slog.Debug("Changing Window to private message", "window", privateMessage.Window.GetID())
	} else {
		s.setActiveWindow(channel.Window)
		slog.Debug("Changing Window to channel", "window", channel.Window.GetID())
	}
	s.updateURL(w, r)
	s.UpdateUI(w, r)
}

func (s *WebClient) handleServer(w http.ResponseWriter, r *http.Request) {
	serverID := r.PathValue("server")
	connection := s.connectionManager.GetConnection(serverID)
	if connection == nil {
		slog.Debug("Invalid change server call, unknown server", "server", serverID)
		s.setActiveWindow(nil)
		s.handleIndex(w, r)
		return
	}
	s.setActiveWindow(connection.GetWindow())
	slog.Debug("Changing Window", "window", s.getActiveWindow().GetID())
	s.handleIndex(w, r)
}

func (s *WebClient) handleChangeServer(w http.ResponseWriter, r *http.Request) {
	serverID := r.PathValue("server")
	connection := s.connectionManager.GetConnection(serverID)
	if connection == nil {
		slog.Debug("Invalid change server call, unknown server", "server", serverID)
		return
	}
	s.setActiveWindow(connection.GetWindow())
	slog.Debug("Changing Window", "window", connection.GetID())
	s.updateURL(w, r)
	s.UpdateUI(w, r)
}

func (s *WebClient) handleInput(w http.ResponseWriter, r *http.Request) {
	type inputValues struct {
		Input string `json:"input"`
	}
	inputData := &inputValues{}
	err := datastar.ReadSignals(r, inputData)
	if err != nil {
		slog.Debug("Error reading input", "error", err)
		return
	}
	input := inputData.Input
	if input == "" {
		return
	}

	s.historyLock.Lock()
	if len(s.inputHistory) == 0 || s.inputHistory[len(s.inputHistory)-1] != input {
		s.inputHistory = append(s.inputHistory, input)
	}
	s.historyPosition = -1
	s.historyLock.Unlock()

	s.commands.Execute(s.connectionManager, s.getActiveWindow(), input)
	s.lock.Lock()
	sse := datastar.NewSSE(w, r)
	var data bytes.Buffer
	err = s.templates.ExecuteTemplate(&data, "EmptyInput.gohtml", nil)
	if err != nil {
		slog.Debug("Error generating template", "error", err)
	}
	err = sse.MergeFragments(data.String())
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		s.lock.Unlock()
		return
	}
	err = sse.MergeSignals([]byte("{input: ''}"))
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		s.lock.Unlock()
		return
	}
	s.lock.Unlock()
	s.UpdateUI(w, r)
}

func (s *WebClient) handleUpload(w http.ResponseWriter, r *http.Request) {
	if s.getActiveWindow() == nil {
		return
	}
	type uploadBody struct {
		Files    []string `json:"files"`
		Mimes    []string `json:"filesMimes"`
		Names    []string `json:"filesNames"`
		FileHost string   `json:"filehost"`
	}
	uploaded := &uploadBody{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(uploaded)
	if err != nil {
		slog.Debug("Error uploading file", "error", err)
		return
	}
	fmt.Println(uploaded.FileHost)
	if len(uploaded.Files) != 1 && len(uploaded.Mimes) != 1 && len(uploaded.Names) != 1 {
		slog.Debug("Error wrong number of files uploaded")
		return
	}
	data, err := base64.StdEncoding.DecodeString(uploaded.Files[0])
	if err != nil {
		slog.Debug("Error decoding file", "error", err)
		return
	}
	if len(uploaded.FileHost) == 0 {
		return
	}
	dataReader := bytes.NewReader(data)
	username, password := s.getActiveWindow().GetServer().GetCredentials()
	if strings.Contains(username, "/") {
		username = strings.Split(username, "/")[0]
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", uploaded.FileHost, dataReader)
	if err != nil {
		slog.Debug("Error creating request file", "error", err)
		return
	}
	req.Header.Set("Content-Type", uploaded.Mimes[0])
	req.Header.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, uploaded.Names[0]))
	req.SetBasicAuth(username, password)
	resp, err := client.Do(req)
	if err != nil {
		slog.Debug("Error uploading file", "error", err)
		return
	}
	if resp.StatusCode != http.StatusCreated {
		defer func() {
			_ = resp.Body.Close()
		}()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Debug("Error reading error", "error", err)
			return
		}
		slog.Debug("File not uploaded", "error", string(body))
		return
	}
	location := resp.Header.Get("location")
	location = strings.TrimPrefix(location, "/uploads")
	slog.Info("File uploaded to bouncer", "file", uploaded.FileHost+location)

	s.lock.Lock()
	sse := datastar.NewSSE(w, r)
	err = sse.MergeSignals([]byte("{files: [], filesMimes: [], filesNames: [], location: \"" + uploaded.FileHost + location + "\"}"))
	if err != nil {
		slog.Debug("Error removing signals", "error", err)
		return
	}
	s.lock.Unlock()
}

func (s *WebClient) handleJoin(w http.ResponseWriter, r *http.Request) {
	if s.getActiveWindow() == nil {
		return
	}
	err := s.getActiveWindow().GetServer().JoinChannel(r.URL.Query().Get("channel"), r.URL.Query().Get("key"))
	if err != nil {
		slog.Debug("Error joining channel", "error", err)
		return
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	sse := datastar.NewSSE(w, r)
	var data bytes.Buffer
	err = s.templates.ExecuteTemplate(&data, "EmptyDialog.gohtml", nil)
	if err != nil {
		slog.Debug("Error generating template", "error", err)
	}
	err = sse.MergeFragments(data.String())
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
}

func (s *WebClient) handlePart(w http.ResponseWriter, r *http.Request) {
	if s.getActiveWindow() == nil {
		return
	}
	err := s.getActiveWindow().GetServer().PartChannel(r.URL.Query().Get("channel"))
	if err != nil {
		slog.Debug("Error parting channel", "error", err)
		return
	}
	s.UpdateUI(w, r)
}

func (s *WebClient) handleNextWindowUp(w http.ResponseWriter, r *http.Request) {
	s.changeWindow(-1)
	s.updateURL(w, r)
	s.UpdateUI(w, r)
}
func (s *WebClient) handleNextWindowDown(w http.ResponseWriter, r *http.Request) {
	s.changeWindow(+1)
	s.updateURL(w, r)
	s.UpdateUI(w, r)
}

func (s *WebClient) updateURL(w http.ResponseWriter, r *http.Request) {
	if s.activeWindow == nil {
		return
	}
	sse := datastar.NewSSE(w, r)
	if s.activeWindow.IsServer() {
		_ = sse.ExecuteScript("window.history.replaceState({}, '', '/s/"+s.activeWindow.GetID()+"')", datastar.WithExecuteScriptAutoRemove(true))
	} else {
		_ = sse.ExecuteScript("window.history.replaceState({}, '', '/s/"+s.activeWindow.GetServer().GetID()+"/"+url.QueryEscape(s.activeWindow.GetName())+"')", datastar.WithExecuteScriptAutoRemove(true))
	}
}

func (s *WebClient) changeWindow(change int) {
	s.listlock.Lock()
	defer s.listlock.Unlock()
	index := slices.IndexFunc(s.serverList.OrderedList, func(item *ServerListItem) bool {
		return item.Window == s.getActiveWindow()
	})
	if len(s.serverList.OrderedList) > 0 {
		if index+change < 0 {
			s.setActiveWindow(s.serverList.OrderedList[len(s.serverList.OrderedList)-1].Window)
		} else if index+change >= len(s.serverList.OrderedList) {
			s.setActiveWindow(s.serverList.OrderedList[0].Window)
		} else {
			s.setActiveWindow(s.serverList.OrderedList[index+change].Window)
		}
	}
}

func (s *WebClient) handleTab(w http.ResponseWriter, r *http.Request) {
	aw := s.getActiveWindow()
	if aw == nil {
		return
	}
	type inputValues struct {
		Input    string `json:"input"`
		Position int    `json:"char"`
	}
	type outputValues struct {
		Input    string `json:"input"`
		Position int    `json:"bs"`
	}
	data := &inputValues{}
	err := datastar.ReadSignals(r, data)
	if err != nil {
		slog.Error("Unable to read signals", "error", err)
		return
	}
	tc := aw.GetTabCompleter()
	input, position := tc.Complete(data.Input, data.Position)
	sse := datastar.NewSSE(w, r)
	output := outputValues{
		Input:    input,
		Position: position,
	}
	dataBytes, err := json.Marshal(output)
	if err != nil {
		slog.Error("Unable to marshal signals", "error", err)
		return
	}
	err = sse.MergeSignals(dataBytes)
	if err != nil {
		slog.Error("Unable to merge signals", "error", err)
		return
	}
}

func (s *WebClient) handleUpdateNicklist(_ http.ResponseWriter, r *http.Request) {
	type showNicklist struct {
		ShowNicklist bool `json:"nicklistshow"`
	}
	sn := &showNicklist{}
	err := datastar.ReadSignals(r, sn)
	if err != nil {
		slog.Debug("Error reading input", "error", err)
		return
	}
	s.conf.UISettings.ShowNicklist = sn.ShowNicklist
}

func (s *WebClient) handleHistoryUp(w http.ResponseWriter, r *http.Request) {
	s.historyLock.Lock()
	defer s.historyLock.Unlock()

	if len(s.inputHistory) == 0 {
		return
	}

	if s.historyPosition == -1 {
		inputData := &inputValues{}
		err := datastar.ReadSignals(r, inputData)
		if err != nil {
			slog.Debug("Error reading input", "error", err)
			return
		}

		if inputData.Input != "" && (len(s.inputHistory) == 0 || s.inputHistory[len(s.inputHistory)-1] != inputData.Input) {
			s.inputHistory = append(s.inputHistory, inputData.Input)
			s.historyPosition = len(s.inputHistory) - 1
		}
	}

	if s.historyPosition == -1 {
		s.historyPosition = len(s.inputHistory) - 1
	} else if s.historyPosition > 0 {
		s.historyPosition--
	}

	inputData := &inputValues{
		Input: s.inputHistory[s.historyPosition],
	}
	sse := datastar.NewSSE(w, r)
	err := sse.MarshalAndMergeSignals(inputData)
	if err != nil {
		slog.Debug("Error merging signals", "error", err)
		return
	}
}

func (s *WebClient) handleHistoryDown(w http.ResponseWriter, r *http.Request) {
	s.historyLock.Lock()
	defer s.historyLock.Unlock()

	if len(s.inputHistory) == 0 || s.historyPosition == -1 {
		return
	}

	if s.historyPosition < len(s.inputHistory)-1 {
		s.historyPosition++

		inputData := &inputValues{
			Input: s.inputHistory[s.historyPosition],
		}
		sse := datastar.NewSSE(w, r)
		err := sse.MarshalAndMergeSignals(inputData)
		if err != nil {
			slog.Debug("Error merging signals", "error", err)
		}
	} else {
		inputData := &inputValues{
			Input: "",
		}
		sse := datastar.NewSSE(w, r)
		err := sse.MarshalAndMergeSignals(inputData)
		if err != nil {
			slog.Debug("Error merging signals", "error", err)
			return
		}
		s.historyPosition = -1
	}
}
