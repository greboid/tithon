// server.go
package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"ircclient/irc"

	"github.com/kouhin/envflag"
)

//go:embed public
var public embed.FS

var (
	port = flag.Int("port", 8080, "Port for the webserver to listen on")
	databaseDirectory = flag.String("db-dir", filepath.Join(".", "database"), "Directory used to store database contents")
)

func GetEmbedOrOSFS(path string, embedFs embed.FS) (fs.FS, error) {
	_, err := os.Stat(path)
	if err == nil {
		return os.DirFS(path), nil
	}
	_, err = embedFs.Open(path)
	if err != nil {
		return nil, err
	}
	staticFiles, err := fs.Sub(embedFs, path)
	if err != nil {
		return nil, err
	}
	return staticFiles, nil
}

func main() {
	if err := envflag.Parse(); err != nil {
		log.Fatalf("Unable to parse flags: %s", err.Error())
	}
	publicFS, err := GetEmbedOrOSFS("public", public)
	if err != nil {
		log.Fatalf("Unable to find web content: %s", err.Error())
	}
	client, err := irc.NewIRCClient(*databaseDirectory)
	if err != nil {
		log.Fatalf("Unable to launch client: %s", err.Error())
	}
	client.Start()
	defer func() {
		_ = client.Stop()
	}()
	router := http.NewServeMux()
	router.Handle("/", http.StripPrefix("/", http.FileServer(http.FS(publicFS))))
	router.HandleFunc("/socket", irc.SocketHandler(client))
	log.Printf("Starting server: http://127.0.0.1:%d", *port)
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: router,
	}
	go func() {
		_ = server.ListenAndServe()
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Unable to shutdown: %s", err.Error())
	}
	log.Print("Finishing server.")
}
