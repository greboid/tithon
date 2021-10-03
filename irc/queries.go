package irc

import (
	"time"
)

type QueryMessage struct {
	Source  string
	time    time.Time
	Message string
}

type Query struct {
	Name           string         `yaml:"name"`
	RecentMessages []QueryMessage `yaml:"-"`
}
