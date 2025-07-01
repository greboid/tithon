package web

import (
	"bytes"
	"encoding/json"
	datastar "github.com/starfederation/datastar/sdk/go"
	"io"
	"log/slog"
	"net/http"
)

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
