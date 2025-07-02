package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetConfigNames(t *testing.T) {
	originalDirName := configDirName
	originalFilename := configFilename
	defer func() {
		configDirName = originalDirName
		configFilename = originalFilename
	}()

	SetConfigNames("custom-dir", "custom-config.yaml")
	assert.Equal(t, "custom-dir", GetConfigDirName())
	assert.Equal(t, "custom-config.yaml", GetConfigFilename())

	SetConfigNames("another-dir", "")
	assert.Equal(t, "another-dir", GetConfigDirName())
	assert.Equal(t, "custom-config.yaml", GetConfigFilename())

	SetConfigNames("", "another-config.yaml")
	assert.Equal(t, "another-dir", GetConfigDirName())
	assert.Equal(t, "another-config.yaml", GetConfigFilename())

	SetConfigNames("", "")
	assert.Equal(t, "another-dir", GetConfigDirName())
	assert.Equal(t, "another-config.yaml", GetConfigFilename())
}

func TestGetUserConfigDir(t *testing.T) {
	originalDirName := configDirName
	defer func() {
		configDirName = originalDirName
	}()

	SetConfigNames("test-dir", "")
	configDir := GetUserConfigDir()

	assert.True(t, strings.HasSuffix(configDir, "test-dir"))
}

func TestGetUserCacheDir(t *testing.T) {
	originalDirName := configDirName
	defer func() {
		configDirName = originalDirName
	}()

	SetConfigNames("test-cache-dir", "")
	cacheDir := GetUserCacheDir()

	assert.True(t, strings.HasSuffix(cacheDir, "test-cache-dir"))
}

func TestDefaultValues(t *testing.T) {
	originalDirName := configDirName
	originalFilename := configFilename
	defer func() {
		configDirName = originalDirName
		configFilename = originalFilename
	}()

	configDirName = "tithon"
	configFilename = "config.yaml"

	assert.Equal(t, "tithon", GetConfigDirName())
	assert.Equal(t, "config.yaml", GetConfigFilename())
}
