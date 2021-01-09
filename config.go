package main

import (
	"errors"
	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
)

// Struct for unmarshaling of opensend TOML configs
type Config struct {
	Receiver ReceiverConfig
	Sender SenderConfig
	Targets map[string]Target
}

// Config section for receiver
type ReceiverConfig struct {
	DestDir string `toml:"destinationDirectory"`
	SkipZeroconf bool
	WorkDir string `toml:"workingDirectory"`
}

// Config section for sender
type SenderConfig struct {
	WorkDir string `toml:"workingDirectory"`
}

type Target struct {
	IP string
}

// Attempt to find config path
func GetConfigPath() string {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
	// Possible config locations
	configLocations := []string{"~/.config/opensend.toml", "/etc/opensend.toml"}
	// For every possible location
	for _, configLocation := range configLocations {
		// Expand path (~ -> home dir and os.ExpandEnv())
		expandedPath := ExpandPath(configLocation)
		// If file does not exist
		if _, err := os.Stat(expandedPath); errors.Is(err, os.ErrNotExist) {
			// Skip
			continue
		}
		// Return path with existing file
		return expandedPath
	}
	// If all else fails, return empty screen
	return ""
}

// Create new config object using values from given path
func NewConfig(path string) *Config {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
	// Create new empty config struct
	newConfig := &Config{}
	// Set config defaults
	newConfig.SetDefaults()
	// If path is provided
	if path != "" {
		// Read file at path
		confData, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatal().Err(err).Msg("Error reading config")
		}
		// Unmarshal config data
		err = toml.Unmarshal(confData, newConfig)
		if err != nil {
			log.Fatal().Err(err).Msg("Error unmarshalling toml")
		}
	}
	// Return new config struct
	return newConfig
}

// Set config defaults
func (config *Config) SetDefaults() {
	// Set destination directory to $HOME/Downloads
	config.Receiver.DestDir = ExpandPath("~/Downloads")
	// Set receiver working directory to $HOME/.opensend
	config.Receiver.WorkDir = ExpandPath("~/.opensend")
	// Set do not skip zeroconf
	config.Receiver.SkipZeroconf = false
	// Set sender working directory to $HOME/.opensend
	config.Sender.WorkDir = ExpandPath("~/.opensend")
	// Set targets to an empty map[string]map[string]string
	config.Targets = map[string]Target{}
}
