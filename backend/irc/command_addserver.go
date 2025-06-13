package irc

import (
	"errors"
	"strconv"
	"strings"
)

type AddServer struct{}

func (c AddServer) GetName() string {
	return "addserver"
}

func (c AddServer) GetHelp() string {
	return "Adds a new server and connects to it. Usage: /addserver hostname[:port] [nickname] [--notls] [--password=serverpassword] [--sasl=username:password]"
}

func (c AddServer) Execute(cm *ConnectionManager, _ *Window, input string) error {
	if input == "" {
		return errors.New("no hostname specified")
	}

	args := strings.Fields(input)
	if len(args) == 0 {
		return errors.New("no hostname specified")
	}

	hostPort := args[0]
	hostname, port := parseHostPort(hostPort)

	nickname := ""
	if len(args) > 1 && !strings.HasPrefix(args[1], "--") {
		nickname = args[1]
	}

	tls := true
	password := ""
	saslLogin := ""
	saslPassword := ""

	for _, arg := range args {
		if arg == "--notls" {
			tls = false
		} else if strings.HasPrefix(arg, "--password=") {
			password = strings.TrimPrefix(arg, "--password=")
		} else if strings.HasPrefix(arg, "--sasl=") {
			saslCreds := strings.TrimPrefix(arg, "--sasl=")
			parts := strings.SplitN(saslCreds, ":", 2)
			if len(parts) == 2 {
				saslLogin = parts[0]
				saslPassword = parts[1]
			}
		}
	}
	if tls && port == -1 {
		port = 6697
	} else {
		port = 6667
	}
	profile := NewProfile(nickname)
	cm.AddConnection("", hostname, port, tls, password, saslLogin, saslPassword, profile, true)

	return nil
}

func parseHostPort(hostPort string) (string, int) {
	parts := strings.SplitN(hostPort, ":", 2)
	hostname := parts[0]
	port := -1

	if len(parts) > 1 {
		if p, err := strconv.Atoi(parts[1]); err == nil {
			port = p
		}
	}

	return hostname, port
}
