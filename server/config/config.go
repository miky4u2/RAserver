package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Version of RAserver
const Version = `1.0.0`

// AppBasePath is the base path of our application
var AppBasePath string

// Settings : Configuration settings
//
var Settings config = config{}

// config type to load and hold configuration settings
//
type config struct {
	BindIP           string   `json:"bindIP"`
	BindPort         string   `json:"bindPort"`
	AllowedIPs       []string `json:"allowedIPs"`
	ValidateAgentTLS bool     `json:"validateAgentTLS"`
	LogFile          string   `json:"logFile"`
	LogToFile        bool     `json:"logToFile"`
}

// Loads configuration settings & set log output to file if required
//
func (c *config) Load() error {
	// work out application base path
	AppBasePath, _ = os.Executable()
	AppBasePath = filepath.Dir(AppBasePath)
	AppBasePath = strings.TrimSuffix(AppBasePath, `bin`)

	filename := filepath.Join(AppBasePath, `conf`, `config.json`)

	configFile, err := os.Open(filename)
	defer configFile.Close()

	if err != nil {
		return err
	}

	jasonParser := json.NewDecoder(configFile)
	err = jasonParser.Decode(c)
	if err != nil {
		return err
	}

	return err
}
