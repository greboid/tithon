package irc

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type AddServer struct{}

func (c AddServer) GetName() string {
	return "addserver" // Using different name to avoid conflicts with existing command
}

func (c AddServer) GetHelp() string {
	return "Adds a new server and connects to it."
}

func (c AddServer) GetUsage() string {
	return GenerateDetailedHelp(c)
}

func (c AddServer) GetArgSpecs() []Argument {
	return []Argument{
		{
			Name:        "hostname",
			Type:        ArgTypeString,
			Required:    true,
			Description: "Server hostname with optional port (e.g., irc.server.com:6667)",
			Validator:   validateNonEmpty,
		},
		{
			Name:        "nickname",
			Type:        ArgTypeNick,
			Required:    false,
			Default:     "",
			Description: "Nickname to use on the server",
		},
	}
}

func (c AddServer) GetFlagSpecs() []Flag {
	return []Flag{
		{
			Name:        "notls",
			Type:        ArgTypeBool,
			Required:    false,
			Default:     false,
			Description: "Disable TLS encryption",
		},
		{
			Name:        "password",
			Short:       "p",
			Type:        ArgTypeString,
			Required:    false,
			Default:     "",
			Description: "Server password",
		},
		{
			Name:        "sasl",
			Short:       "s",
			Type:        ArgTypeString,
			Required:    false,
			Default:     "",
			Description: "SASL authentication credentials (username:password)",
		},
	}
}

func (c AddServer) Execute(cm *ServerManager, _ *Window, input string) error {
	parsed, err := Parse(c, input)
	if err != nil {
		return fmt.Errorf("argument parsing error: %w", err)
	}

	hostname, err := parsed.GetArgString("hostname")
	if err != nil {
		return fmt.Errorf("failed to get hostname: %w", err)
	}

	nickname, err := parsed.GetArgString("nickname")
	if err != nil {
		return fmt.Errorf("failed to get nickname: %w", err)
	}

	notls, err := parsed.GetFlagBool("notls")
	if err != nil {
		return fmt.Errorf("failed to get notls flag: %w", err)
	}

	password, err := parsed.GetFlagString("password")
	if err != nil {
		return fmt.Errorf("failed to get password flag: %w", err)
	}

	saslCreds, err := parsed.GetFlagString("sasl")
	if err != nil {
		return fmt.Errorf("failed to get sasl flag: %w", err)
	}

	host, port := parseHostPort(hostname)

	if port == -1 {
		if notls {
			port = 6667
		} else {
			port = 6697
		}
	}

	var saslLogin, saslPassword string
	if saslCreds != "" {
		parts := splitSASLCredentials(saslCreds)
		if len(parts) != 2 {
			return errors.New("SASL credentials must be in format username:password")
		}
		saslLogin = parts[0]
		saslPassword = parts[1]
	}

	profile := NewProfile(nickname)
	cm.AddConnection("", host, port, !notls, password, saslLogin, saslPassword, profile, false)

	return nil
}

func splitSASLCredentials(creds string) []string {
	for i, char := range creds {
		if char == ':' {
			return []string{creds[:i], creds[i+1:]}
		}
	}
	return []string{creds}
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
