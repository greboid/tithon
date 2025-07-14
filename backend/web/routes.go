package web

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	uniqueid "github.com/albinj12/unique-id"
	"github.com/greboid/tithon/config"
	"github.com/greboid/tithon/irc"
	"github.com/greboid/tithon/services"
	datastar "github.com/starfederation/datastar/sdk/go"
	"html"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

func (s *WebClient) addRoutes(mux *http.ServeMux) {
	static, usercss := s.setupFsAndGetWatchers()
	mux.HandleFunc("GET /static/", s.handleStatic(static))
	mux.HandleFunc("GET /static/user.css", s.handleUserCSS(usercss))
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

func (s *WebClient) handleIndex(w http.ResponseWriter, _ *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pendingUpdate.Store(true)
	err := s.templates.ExecuteTemplate(w, "Index.gohtml", s.settingsService.GetSettingsData().Version)
	if err != nil {
		slog.Debug("Error serving index", "error", err)
		return
	}
}

func (s *WebClient) handleStatic(static fs.FS) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s.noCacheMiddleware(http.StripPrefix("/static/", http.FileServer(http.FS(static)))).ServeHTTP(w, r)
	}
}

func (s *WebClient) handleUserCSS(usercss string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, usercss)
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
		case notification := <-s.notificationService.GetNotificationChannel():
			slog.Debug("Sending notification", "notification", notification)
			err := datastar.NewSSE(w, r).ExecuteScript(
				fmt.Sprintf(`notify("%s", "%s", %t, %t, "")`,
					html.EscapeString(notification.Title),
					html.EscapeString(notification.Text),
					notification.Popup,
					notification.Sound,
				), datastar.WithExecuteScriptAutoRemove(true))
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

	err := s.templates.ExecuteTemplate(&data, "SettingsPage.gohtml", nil)
	if err != nil {
		slog.Debug("Error generating template", "error", err)
	}
	err = s.templates.ExecuteTemplate(&data, "SettingsContent.gohtml", s.settingsService.GetFromConfig())
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

	err := s.templates.ExecuteTemplate(&data, "SettingsContent.gohtml", s.settingsService.GetSettingsData())
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

	settingsData := s.settingsService.GetSettingsData()
	settingsData.Servers = slices.DeleteFunc(settingsData.Servers, func(server config.Server) bool {
		return server.ID == serverID
	})

	err := s.templates.ExecuteTemplate(&data, "SettingsContent.gohtml", s.settingsService.GetSettingsData())
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

	settingsData := s.settingsService.GetSettingsData()
	index := slices.IndexFunc(settingsData.Servers, func(server config.Server) bool {
		return server.ID == serverID
	})

	if index == -1 {
		slog.Debug("Unknown server specified", "WebClient ID", serverID)
		return
	}

	s.connectionManager.AddConnection(
		settingsData.Servers[index].ID,
		settingsData.Servers[index].Hostname,
		settingsData.Servers[index].Port,
		settingsData.Servers[index].TLS,
		settingsData.Servers[index].Password,
		settingsData.Servers[index].SASLLogin,
		settingsData.Servers[index].SASLPassword,
		irc.NewProfile(settingsData.Servers[index].Profile.Nickname),
		true,
	)

	var data bytes.Buffer
	err := s.templates.ExecuteTemplate(&data, "SettingsContent.gohtml", s.settingsService.GetSettingsData())
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
	settingsData := s.settingsService.GetSettingsData()
	settingsData.Servers = append(settingsData.Servers, config.Server{
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
	err = s.templates.ExecuteTemplate(&data, "SettingsContent.gohtml", s.settingsService.GetSettingsData())
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

	settingsData := s.settingsService.GetSettingsData()
	index := slices.IndexFunc(settingsData.Servers, func(server config.Server) bool {
		return server.ID == serverID
	})
	con := settingsData.Servers[index]
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

	settingsData := s.settingsService.GetSettingsData()
	for i := range settingsData.Servers {
		if settingsData.Servers[i].ID == id {
			settingsData.Servers[i].Hostname = hostname
			settingsData.Servers[i].Port = portInt
			settingsData.Servers[i].TLS = tlsBool
			settingsData.Servers[i].Password = password
			settingsData.Servers[i].SASLLogin = sasllogin
			settingsData.Servers[i].SASLPassword = saslpassword
			settingsData.Servers[i].Profile.Nickname = nickname
			settingsData.Servers[i].AutoConnect = autoConnectBool
		}
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	sse := datastar.NewSSE(w, r)
	var data bytes.Buffer
	err = s.templates.ExecuteTemplate(&data, "SettingsContent.gohtml", s.settingsService.GetSettingsData())
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

	settingsData := s.settingsService.GetSettingsData()
	settingsData.TimestampFormat = timestampFormat
	settingsData.ShowNicklist = showNicklist
	settingsData.Theme = theme

	err := s.settingsService.SaveSettingsToConfig()
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

	s.inputHistoryService.AddToHistory(input)

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
	slog.Debug("Uploading file")
	if s.getActiveWindow() == nil {
		return
	}
	type uploadBody struct {
		Files    []string `json:"files"`
		Mimes    []string `json:"filesMimes"`
		Names    []string `json:"filesNames"`
		FileHost string   `json:"filehost"`
		Location string   `json:"location"`
		Input    string   `json:"input"`
		Position int      `json:"position"`
	}
	uploaded := &uploadBody{}
	err := json.NewDecoder(r.Body).Decode(uploaded)
	if err != nil {
		slog.Debug("Error uploading file", "error", err)
		return
	}
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
	var username, password, apiKey string
	if s.conf.UISettings.UploadURL != "" {
		if s.conf.UISettings.UploadAPIKey != "" {
			apiKey = s.conf.UISettings.UploadAPIKey
		}
	} else {
		username, password = s.getActiveWindow().GetServer().GetCredentials()
		if strings.Contains(username, "/") {
			username = strings.Split(username, "/")[0]
		}
	}
	client := &http.Client{}
	method := "POST"
	if s.conf.UISettings.UploadMethod != "" {
		method = s.conf.UISettings.UploadMethod
	}
	req, err := http.NewRequest(method, uploaded.FileHost, dataReader)
	if err != nil {
		slog.Debug("Error creating request file", "error", err)
		return
	}
	if len(uploaded.Names) > 0 {
		req.Header.Set("Content-Type", uploaded.Mimes[0])
	}
	if len(uploaded.Names) > 0 {
		req.Header.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, uploaded.Names[0]))
	}
	if apiKey == "" {
		req.SetBasicAuth(username, password)
	} else {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	}
	resp, err := client.Do(req)
	if err != nil {
		slog.Debug("Error uploading file", "error", err)
		return
	}
	if resp.StatusCode < 199 && resp.StatusCode > 299 {
		defer func() {
			_ = resp.Body.Close()
		}()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Debug("Error reading error", "error", err)
			return
		}
		slog.Debug("File not uploaded", "status", resp.StatusCode, "error", string(body))
		return
	}
	location := resp.Header.Get("location")
	if location == "" {
		locationData, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Debug("Unable to read body", "error", err)
		} else {
			uploaded.Location = string(locationData)
		}
	} else {
		location = strings.TrimPrefix(location, "/uploads")
		uploaded.Location = uploaded.FileHost + location
	}
	uploaded.Files = []string{}
	uploaded.Mimes = []string{}
	uploaded.Names = []string{}
	uploaded.Input = uploaded.Input[:uploaded.Position] + uploaded.Location + uploaded.Input[uploaded.Position:]
	slog.Debug("File uploaded", "location", uploaded.Location)
	s.lock.Lock()
	sse := datastar.NewSSE(w, r)
	err = sse.MarshalAndMergeSignals(uploaded)
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
	activeWindow := s.getActiveWindow()
	if activeWindow == nil {
		return
	}
	sse := datastar.NewSSE(w, r)
	if activeWindow.IsServer() {
		_ = sse.ExecuteScript("window.history.replaceState({}, '', '/s/"+activeWindow.GetID()+"')", datastar.WithExecuteScriptAutoRemove(true))
	} else {
		_ = sse.ExecuteScript("window.history.replaceState({}, '', '/s/"+activeWindow.GetServer().GetID()+"/"+url.QueryEscape(activeWindow.GetName())+"')", datastar.WithExecuteScriptAutoRemove(true))
	}
}

func (s *WebClient) changeWindow(change int) {
	serverList := s.getServerList()
	index := slices.IndexFunc(serverList.OrderedList, func(item *services.ServerListItem) bool {
		return item.Window == s.getActiveWindow()
	})
	if len(serverList.OrderedList) > 0 {
		if index+change < 0 {
			s.setActiveWindow(serverList.OrderedList[len(serverList.OrderedList)-1].Window)
		} else if index+change >= len(serverList.OrderedList) {
			s.setActiveWindow(serverList.OrderedList[0].Window)
		} else {
			s.setActiveWindow(serverList.OrderedList[index+change].Window)
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
	if s.inputHistoryService.GetHistoryLength() == 0 {
		return
	}

	var currentInput string
	if s.inputHistoryService.GetCurrentPosition() == -1 {
		inputData := &inputValues{}
		err := datastar.ReadSignals(r, inputData)
		if err != nil {
			slog.Debug("Error reading input", "error", err)
			return
		}
		currentInput = inputData.Input
	}

	historyItem := s.inputHistoryService.NavigateUp(currentInput)
	inputData := &inputValues{
		Input: historyItem,
	}
	sse := datastar.NewSSE(w, r)
	err := sse.MarshalAndMergeSignals(inputData)
	if err != nil {
		slog.Debug("Error merging signals", "error", err)
		return
	}
}

func (s *WebClient) handleHistoryDown(w http.ResponseWriter, r *http.Request) {
	if s.inputHistoryService.GetHistoryLength() == 0 || s.inputHistoryService.GetCurrentPosition() == -1 {
		return
	}

	historyItem := s.inputHistoryService.NavigateDown()
	inputData := &inputValues{
		Input: historyItem,
	}
	sse := datastar.NewSSE(w, r)
	err := sse.MarshalAndMergeSignals(inputData)
	if err != nil {
		slog.Debug("Error merging signals", "error", err)
	}
}
