package web

import (
	"errors"
	"github.com/fsnotify/fsnotify"
	"github.com/kirsle/configdir"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
)

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

func (s *WebClient) setupFsAndGetWatchers() (fs.FS, string) {
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
	return static, usercss
}
