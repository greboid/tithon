package config

import (
	"fmt"
	"os"
	"path/filepath"

	uniqueid "github.com/albinj12/unique-id"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"time"
)

var (
	configDirName  = "tithon"
	configFilename = "config.yaml"
)

func SetConfigNames(dirName, filename string) {
	if dirName != "" {
		configDirName = dirName
	}
	if filename != "" {
		configFilename = filename
	}
}

type Config struct {
	instance      Provider
	Servers       []Server      `yaml:"servers" validate:"dive"`
	UISettings    UISettings    `yaml:"ui_settings" validate:"required"`
	Notifications Notifications `yaml:"notifications"`
}

func NewConfig(provider Provider) *Config {
	return &Config{
		instance: provider,
	}
}

func GetUserConfigDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	return filepath.Join(configDir, GetConfigDirName())
}

func GetUserCacheDir() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = "."
	}
	return filepath.Join(cacheDir, GetConfigDirName())
}

func GetConfigDirName() string {
	return configDirName
}

func GetConfigFilename() string {
	return configFilename
}

type Server struct {
	Hostname     string  `yaml:"hostname" validate:"required,hostname_rfc1123|ip"`
	Port         int     `yaml:"port" validate:"min=1,max=65535"`
	TLS          bool    `yaml:"tls"`
	Password     string  `yaml:"password"`
	SASLLogin    string  `yaml:"sasl_login,omitempty" validate:"required_with=SASLPassword"`
	SASLPassword string  `yaml:"sasl_password,omitempty" validate:"required_with=SASLLogin"`
	Profile      Profile `yaml:"profile" validate:"required"`
	ID           string  `yaml:"id"`
	AutoConnect  bool    `yaml:"auto_connect"`
}

type UISettings struct {
	TimestampFormat string `yaml:"timestamp_format"`
	ShowNicklist    bool   `yaml:"show_nicklist"`
	Theme           string `yaml:"theme" validate:"omitempty,oneof=light dark auto"`
	UploadURL       string `yaml:"upload-url,omitempty" validate:"omitempty,http_url"`
	UploadAPIKey    string `yaml:"upload-api-key,omitempty"`
	UploadMethod    string `yaml:"upload-method,omitempty" validate:"omitempty,oneof=POST PUT post put"`
}

type Profile struct {
	Nickname string `yaml:"nickname" validate:"required,min=1,max=30"`
}

type Notifications struct {
	Triggers []NotificationTrigger `yaml:"triggers"`
}

type NotificationTrigger struct {
	Network string `yaml:"network"`
	Source  string `yaml:"source"`
	Nick    string `yaml:"nick"`
	Message string `yaml:"message"`
	Sound   bool   `yaml:"sound" validate:"required_without Popup"`
	Popup   bool   `yaml:"popup" validate:"required_without Sound"`
}

func (c *Config) Load() error {
	slog.Debug("Loading config")
	if err := c.instance.Load(c); err != nil {
		return err
	}

	c.applyDefaults()

	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(c); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}
	return nil
}

func (c *Config) Save() error {
	return c.instance.Save(c)
}

// applyDefaults sets default values for optional configuration fields
func (c *Config) applyDefaults() {
	// Set default timestamp format
	if c.UISettings.TimestampFormat == "" {
		c.UISettings.TimestampFormat = time.TimeOnly
	}

	// Set default theme
	if c.UISettings.Theme == "" {
		c.UISettings.Theme = "auto"
	}

	// Generate IDs for servers that don't have them
	for i := range c.Servers {
		if c.Servers[i].ID == "" {
			id, _ := uniqueid.Generateid("a", 5, "s")
			c.Servers[i].ID = id
		}

		// Set default port if not specified
		if c.Servers[i].Port == 0 {
			if c.Servers[i].TLS {
				c.Servers[i].Port = 6697
			} else {
				c.Servers[i].Port = 6667
			}
		}
	}
}
