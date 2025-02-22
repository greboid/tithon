#!/bin/bash
go run github.com/a-h/templ/cmd/templ@latest generate --watch --proxy=http://localhost:8080 --cmd "go run . --debug --port 8080"
