package web

import (
	"github.com/greboid/ircclient/irc"
	"github.com/greboid/ircclient/web/templates"
	datastar "github.com/starfederation/datastar/sdk/go"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strconv"
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
	mux.HandleFunc("GET /closeDialog", s.handleCloseDialog)
	mux.HandleFunc("GET /addServer", s.handleAddServer)
	mux.HandleFunc("GET /changeWindow/{server}", s.handleChangeServer)
	mux.HandleFunc("GET /changeWindow/{server}/{channel}", s.handleChangeChannel)
	mux.HandleFunc("GET /input", s.handleInput)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	err := templates.Loading().Render(r.Context(), w)
	if err != nil {
		slog.Debug("Error serving index", "error", err)
		return
	}
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()
	sse := datastar.NewSSE(w, r)
	err := sse.MergeFragmentTempl(templates.Index(), func(options *datastar.MergeFragmentOptions) {
		options.Selector = "body"
		options.MergeMode = datastar.FragmentMergeModeInner
	})
	if err != nil {
		slog.Debug("Error serving ready", "error", err)
	}
	err = sse.ExecuteScript("window.history.pushState({}, '', '/#/')")
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
	s.UpdateUI(w, r)
}

func (s *Server) UpdateUI(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)
	err := sse.MergeFragmentTempl(templates.ServerList(s.connectionManager.GetConnections()))
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		s.lock.Unlock()
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
		s.lock.Unlock()
		return
	}

}

func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			slog.Debug("Client connection closed")
			return
		case <-ticker.C:
			s.lock.Lock()
			s.UpdateUI(w, r)
			s.lock.Unlock()
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

func (s *Server) handleCloseDialog(w http.ResponseWriter, r *http.Request) {
	s.lock.Lock()
	defer s.lock.Unlock()
	sse := datastar.NewSSE(w, r)
	err := sse.MergeFragmentTempl(templates.EmptyDialog(), func(options *datastar.MergeFragmentOptions) {
		options.Selector = "#dialog"
	})
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
	err = sse.ExecuteScript("window.history.pushState({}, '', '/#/')")
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
	id := s.connectionManager.AddConnection(hostname, portInt, tlsBool, sasllogin, saslpassword, irc.NewProfile(nickname))
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
	s.lock.Lock()
	defer s.lock.Unlock()
	sse := datastar.NewSSE(w, r)
	server := s.connectionManager.GetConnection(s.activeServer)
	var channel *irc.Channel
	if server == nil {
		channel = nil
	} else {
		channel = server.GetChannel(s.activeWindow)
	}
	err := sse.MergeFragmentTempl(templates.GetWindow(server, channel))
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
}

func (s *Server) handleChangeServer(w http.ResponseWriter, r *http.Request) {
	s.activeServer = r.PathValue("server")
	s.activeWindow = ""
	slog.Debug("Changing Server", "server", s.activeServer)
	s.lock.Lock()
	defer s.lock.Unlock()
	sse := datastar.NewSSE(w, r)
	server := s.connectionManager.GetConnection(s.activeServer)
	var channel *irc.Channel
	if server == nil {
		channel = nil
	} else {
		channel = server.GetChannel(s.activeWindow)
	}
	err := sse.MergeFragmentTempl(templates.GetWindow(server, channel))
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
}

func (s *Server) handleInput(w http.ResponseWriter, r *http.Request) {
	input := r.URL.Query().Get("input")
	if input == "" {
		return
	}
	s.connectionManager.GetConnection(s.activeServer).SendMessage(s.activeWindow, input)
	sse := datastar.NewSSE(w, r)
	err := sse.MergeFragmentTempl(templates.EmptyInput())
	if err != nil {
		slog.Debug("Error merging fragments", "error", err)
		return
	}
	s.UpdateUI(w, r)
}
