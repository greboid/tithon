package web

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/enescakir/emoji"
	"github.com/greboid/ircclient/irc"
	"github.com/greboid/ircclient/web/templates"
	datastar "github.com/starfederation/datastar/sdk/go"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func (s *Server) addRoutes(mux *http.ServeMux) {
	var static fs.FS
	if stat, err := os.Stat("./web/static"); err == nil && stat.IsDir() {
		slog.Debug("Using on disk static resources")
		static = os.DirFS("./web/static")
	} else {
		slog.Debug("Using on embedded static resources")
		static, _ = fs.Sub(staticFS, "static")
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static))))
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("GET /ready", s.handleReady)
	mux.HandleFunc("GET /update", s.handleUpdate)
	mux.HandleFunc("GET /showSettings", s.handleShowSettings)
	mux.HandleFunc("GET /showAddServer", s.handleShowAddServer)
	mux.HandleFunc("GET /addServer", s.handleAddServer)
	mux.HandleFunc("GET /changeWindow/{server}", s.handleChangeServer)
	mux.HandleFunc("GET /changeWindow/{server}/{channel}", s.handleChangeChannel)
	mux.HandleFunc("GET /input", s.handleInput)
	mux.HandleFunc("POST /upload", s.handleUpload)
	mux.HandleFunc("GET /join", s.handleJoin)
	mux.HandleFunc("GET /part", s.handlePart)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	err := templates.Base(true, s.connectionManager.GetConnections(), s.activeWindow).Render(r.Context(), w)
	if err != nil {
		slog.Debug("Error serving index", "error", err)
		return
	}
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	sse := datastar.NewSSE(w, r)
	err := sse.MergeFragmentTempl(templates.Index(s.connectionManager.GetConnections(), s.activeWindow))
	if err != nil {
		slog.Debug("Error serving ready", "error", err)
	}
	err = sse.ExecuteScript("window.history.pushState({}, '', '/#/')")
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		s.lock.Unlock()
		return
	}
	s.lock.Unlock()
	s.UpdateUI(w, r)
}

func (s *Server) UpdateUI(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()
	sse := datastar.NewSSE(w, r)
	activeID := ""
	if s.activeWindow == "nil" {
		activeID = s.activeServer
	} else {
		activeID = s.activeWindow
	}
	err := sse.MergeFragmentTempl(templates.ServerList(s.connectionManager.GetConnections(), activeID))
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
	server := s.connectionManager.GetConnection(s.activeServer)
	var channel *irc.Channel
	if server == nil {
		channel = nil
	} else {
		channel = server.GetChannel(s.activeWindow)
	}
	err = sse.MergeFragmentTempl(templates.GetWindow(server, channel))
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
	if s.connectionManager.GetConnection(s.activeServer) == nil {
		return
	}
	err = sse.MergeSignals([]byte(fmt.Sprintf("{filehost: '%s'}",
		s.connectionManager.GetConnection(s.activeServer).GetFileHost())))
	if err != nil {
		slog.Debug("Error merging signals", "error", err)
		return
	}
}

func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			slog.Debug("Client connection closed")
			return
		case <-ticker.C:
			s.UpdateUI(w, r)
		}
	}
}

func (s *Server) handleShowSettings(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()
	sse := datastar.NewSSE(w, r)
	slog.Debug("Showing settings")
	err := sse.MergeFragmentTempl(templates.SettingsPage(), func(options *datastar.MergeFragmentOptions) {
		options.Selector = "#dialog"
	})
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
	err = sse.ExecuteScript("window.history.pushState({}, '', '/#/settings')")
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
}

func (s *Server) handleShowAddServer(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()
	sse := datastar.NewSSE(w, r)
	slog.Debug("Showing settings")
	err := sse.MergeFragmentTempl(templates.AddServerPage(), func(options *datastar.MergeFragmentOptions) {
		options.Selector = "#dialog"
	})
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
	err = sse.ExecuteScript("window.history.pushState({}, '', '/#/addserver')")
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
}

func (s *Server) handleAddServer(w http.ResponseWriter, r *http.Request) {
	hostname := r.URL.Query().Get("hostname")
	port := r.URL.Query().Get("port")
	portInt, err := strconv.Atoi(port)
	if err != nil {
		//TODO: Handle error
		portInt = 6667
	}
	tls := r.URL.Query().Get("tls")
	tlsBool, err := strconv.ParseBool(tls)
	if err != nil {
		//TODO: Handle error
		tlsBool = true
	}
	nickname := r.URL.Query().Get("nickname")
	sasllogin := r.URL.Query().Get("sasllogin")
	saslpassword := r.URL.Query().Get("saslpassword")
	password := r.URL.Query().Get("password")
	id := s.connectionManager.AddConnection(hostname, portInt, tlsBool, password, sasllogin, saslpassword, irc.NewProfile(nickname))
	go func() {
		s.connectionManager.GetConnection(id).Connect()
	}()
	s.lock.Lock()
	defer s.lock.Unlock()
	sse := datastar.NewSSE(w, r)
	err = sse.MergeFragmentTempl(templates.EmptyDialog(), func(options *datastar.MergeFragmentOptions) {
		options.Selector = "#dialog"
	})
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
}

func (s *Server) handleChangeChannel(w http.ResponseWriter, r *http.Request) {
	s.activeServer = r.PathValue("server")
	s.activeWindow = r.PathValue("channel")
	slog.Debug("Changing Window", "server", s.activeServer, "channel", s.activeWindow)
	s.UpdateUI(w, r)
}

func (s *Server) handleChangeServer(w http.ResponseWriter, r *http.Request) {
	s.activeServer = r.PathValue("server")
	s.activeWindow = ""
	slog.Debug("Changing Server", "server", s.activeServer)
	s.UpdateUI(w, r)
}

func (s *Server) handleInput(w http.ResponseWriter, r *http.Request) {
	input := r.URL.Query().Get("input")
	if input == "" {
		return
	}
	input = emoji.Parse(input)
	activeServer := s.connectionManager.GetConnection(s.activeServer)
	var activeWindow *irc.Channel
	if activeServer != nil {
		activeWindow = activeServer.GetChannel(s.activeWindow)
	}
	s.commands.Execute(s.connectionManager, activeServer, activeWindow, input)
	s.lock.Lock()
	sse := datastar.NewSSE(w, r)
	err := sse.MergeFragmentTempl(templates.EmptyInput())
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		s.lock.Unlock()
		return
	}
	s.lock.Unlock()
	s.UpdateUI(w, r)
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if s.connectionManager.GetConnection(s.activeServer) == nil {
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
	username, password := s.connectionManager.GetConnection(s.activeServer).GetCredentials()
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

func (s *Server) handleJoin(w http.ResponseWriter, r *http.Request) {
	if s.connectionManager.GetConnection(s.activeServer) == nil {
		return
	}
	err := s.connectionManager.GetConnection(s.activeServer).JoinChannel(r.URL.Query().Get("channel"), r.URL.Query().Get("key"))
	if err != nil {
		slog.Debug("Error joining channel", "error", err)
		return
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	sse := datastar.NewSSE(w, r)
	err = sse.MergeFragmentTempl(templates.EmptyDialog())
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
}

func (s *Server) handlePart(w http.ResponseWriter, r *http.Request) {
	if s.connectionManager.GetConnection(s.activeServer) == nil {
		return
	}
	err := s.connectionManager.GetConnection(s.activeServer).PartChannel(r.URL.Query().Get("channel"))
	if err != nil {
		slog.Debug("Error parting channel", "error", err)
		return
	}
	s.UpdateUI(w, r)
}
