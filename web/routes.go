package web

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/greboid/ircclient/irc"
	datastar "github.com/starfederation/datastar/sdk/go"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
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

func (s *Server) addRoutes(mux *http.ServeMux) {
	var static fs.FS
	if stat, err := os.Stat("./web/static"); err == nil && stat.IsDir() {
		slog.Debug("Using on disk static resources")
		static = os.DirFS("./web/static")
	} else {
		slog.Debug("Using on embedded static resources")
		static, _ = fs.Sub(staticFS, "static")
	}
	var allTemplates fs.FS
	if stat, err := os.Stat("./web/templates"); err == nil && stat.IsDir() {
		slog.Debug("Using on disk templates")
		allTemplates = os.DirFS("./web/templates")
	} else {
		slog.Debug("Using on embedded templates")
		allTemplates, _ = fs.Sub(templateFS, "templates")
	}
	allParsedTemplates, err := template.New("").Funcs(getTemplateFuncs()).ParseFS(allTemplates, "*.gohtml")
	if err != nil {
		slog.Error("Error parsing templates", "error", err)
		panic("Unable to load templates")
	}
	s.templates = allParsedTemplates
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static))))
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("GET /update", s.handleUpdate)
	mux.HandleFunc("GET /showSettings", s.handleShowSettings)
	mux.HandleFunc("GET /showAddServer", s.handleShowAddServer)
	mux.HandleFunc("GET /addServer", s.handleAddServer)
	mux.HandleFunc("GET /changeWindow/{server}", s.handleChangeServer)
	mux.HandleFunc("GET /changeWindow/{server}/{channel}", s.handleChangeChannel)
	mux.HandleFunc("GET /s/{server}", s.handleServer)
	mux.HandleFunc("GET /s/{server}/{channel}", s.handleChannel)
	mux.HandleFunc("GET /input", s.handleInput)
	mux.HandleFunc("POST /upload", s.handleUpload)
	mux.HandleFunc("GET /join", s.handleJoin)
	mux.HandleFunc("GET /part", s.handlePart)
}

func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()
	err := s.templates.ExecuteTemplate(w, "Index.gohtml", nil)
	if err != nil {
		slog.Debug("Error serving index", "error", err)
		return
	}
}

func (s *Server) UpdateUI(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()
	sse := datastar.NewSSE(w, r)
	var data bytes.Buffer
	s.outputTemplate(&data, "Serverlist.gohtml", map[string]any{
		"Connections":  s.connectionManager.GetConnections(),
		"ActiveWindow": s.activeWindow,
	})
	if s.getActiveWindow() == nil {
		s.outputTemplate(&data, "WindowInfo.gohtml", "")
		s.outputTemplate(&data, "Messages.gohtml", nil)
		s.outputTemplate(&data, "Nicklist.gohtml", nil)
	} else {
		s.outputTemplate(&data, "WindowInfo.gohtml", s.getActiveWindow().GetTitle())
		s.outputTemplate(&data, "Messages.gohtml", s.getActiveWindow().GetMessages())
		s.outputTemplate(&data, "Nicklist.gohtml", s.getActiveWindow().GetUsers())
	}
	err := sse.MergeFragments(data.String())
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
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

func (s *Server) outputTemplate(wr io.Writer, name string, data any) {
	err := s.templates.ExecuteTemplate(wr, name, data)
	if err != nil {
		slog.Debug("Error generating name", "error", err)
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
	var data bytes.Buffer
	err := s.templates.ExecuteTemplate(&data, "SettingsPage.gohtml", nil)
	if err != nil {
		slog.Debug("Error generating template", "error", err)
	}
	err = sse.MergeFragments(data.String(), func(options *datastar.MergeFragmentOptions) {
		options.Selector = "#dialog"
	})
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
	var data bytes.Buffer
	err := s.templates.ExecuteTemplate(&data, "AddServerPage.gohtml", nil)
	if err != nil {
		slog.Debug("Error generating template", "error", err)
	}
	err = sse.MergeFragments(data.String(), func(options *datastar.MergeFragmentOptions) {
		options.Selector = "#dialog"
	})
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
	s.connectionManager.AddConnection(hostname, portInt, tlsBool, password, sasllogin, saslpassword, irc.NewProfile(nickname), true)
	s.lock.Lock()
	defer s.lock.Unlock()
	sse := datastar.NewSSE(w, r)
	var data bytes.Buffer
	err = s.templates.ExecuteTemplate(&data, "EmptyDialog.gohtml", nil)
	if err != nil {
		slog.Debug("Error generating template", "error", err)
	}
	err = sse.MergeFragments(data.String(), func(options *datastar.MergeFragmentOptions) {
		options.Selector = "#dialog"
	})
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
}

func (s *Server) handleChannel(w http.ResponseWriter, r *http.Request) {
	serverID := r.PathValue("server")
	channelID := r.PathValue("channel")
	connection := s.connectionManager.GetConnection(serverID)
	if connection == nil {
		slog.Debug("Invalid change channel call, unknown server", "server", serverID)
		return
	}
	channel := connection.GetChannel(channelID)
	if channel == nil {
		slog.Debug("Invalid change channel call, unknown channel", "server", serverID, "channel", channelID)
		return
	}
	s.setActiveWindow(channel.Window)
	slog.Debug("Changing Window", "window", channel.Window.GetID())
	s.handleIndex(w, r)
}

func (s *Server) handleChangeChannel(w http.ResponseWriter, r *http.Request) {
	serverID := r.PathValue("server")
	channelID := r.PathValue("channel")
	connection := s.connectionManager.GetConnection(serverID)
	if connection == nil {
		slog.Debug("Invalid change channel call, unknown server", "server", serverID)
		return
	}
	channel := connection.GetChannel(channelID)
	if channel == nil {
		slog.Debug("Changing Window", "window", s.getActiveWindow().GetID())
		return
	}
	s.setActiveWindow(channel.Window)
	slog.Debug("Changing Window", "window", channel.Window.GetID())
	sse := datastar.NewSSE(w, r)
	_ = sse.ExecuteScript("window.history.replaceState({}, '', '/s/"+serverID+"/"+channelID+"')", datastar.WithExecuteScriptAutoRemove(true))
	s.UpdateUI(w, r)
}

func (s *Server) handleServer(w http.ResponseWriter, r *http.Request) {
	serverID := r.PathValue("server")
	connection := s.connectionManager.GetConnection(serverID)
	if connection == nil {
		slog.Debug("Invalid change server call, unknown server", "server", serverID)
		return
	}
	s.setActiveWindow(connection.Window)
	slog.Debug("Changing Window", "window", s.getActiveWindow().GetID())
	s.handleIndex(w, r)
}

func (s *Server) handleChangeServer(w http.ResponseWriter, r *http.Request) {
	serverID := r.PathValue("server")
	connection := s.connectionManager.GetConnection(serverID)
	if connection == nil {
		slog.Debug("Invalid change server call, unknown server", "server", serverID)
		return
	}
	s.setActiveWindow(connection.Window)
	slog.Debug("Changing Window", "window", connection.Window.GetID())
	sse := datastar.NewSSE(w, r)
	_ = sse.ExecuteScript("window.history.replaceState({}, '', '/s/"+serverID+"')", datastar.WithExecuteScriptAutoRemove(true))

	s.UpdateUI(w, r)
}

func (s *Server) handleInput(w http.ResponseWriter, r *http.Request) {
	input := r.URL.Query().Get("input")
	if input == "" {
		return
	}
	s.commands.Execute(s.connectionManager, s.getActiveWindow(), input)
	s.lock.Lock()
	sse := datastar.NewSSE(w, r)
	var data bytes.Buffer
	err := s.templates.ExecuteTemplate(&data, "EmptyInput.gohtml", nil)
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

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) handleJoin(w http.ResponseWriter, r *http.Request) {
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
	err = s.templates.ExecuteTemplate(&data, "EmptyDialog.gotpl", nil)
	if err != nil {
		slog.Debug("Error generating template", "error", err)
	}
	err = sse.MergeFragments(data.String())
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
}

func (s *Server) handlePart(w http.ResponseWriter, r *http.Request) {
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
