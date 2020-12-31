package main

import (
	"errors"
	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
)

type Config struct {
	Receiver ReceiverConfig
	Sender   SenderConfig
	Targets  map[string]map[string]string
}

type ReceiverConfig struct {
	DestDir      string `toml:"destinationDirectory"`
	SkipZeroconf bool
	WorkDir      string `toml:"workingDirectory"`
}

type SenderConfig struct {
	WorkDir string `toml:"workingDirectory"`
}

func GetConfigPath() string {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
	configLocations := []string{"~/.config/opensend.toml", "/etc/opensend.toml"}
	for _, configLocation := range configLocations {
		expandedPath := ExpandPath(configLocation)
		if _, err := os.Stat(expandedPath); errors.Is(err, os.ErrNotExist) {
			continue
		}
		return expandedPath
	}
	return ""
}

func NewConfig(path string) *Config {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
	newConfig := &Config{}
	newConfig.SetDefaults()
	if path != "" {
		confData, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatal().Err(err).Msg("Error reading config")
		}
		err = toml.Unmarshal(confData, newConfig)
		if err != nil {
			log.Fatal().Err(err).Msg("Error unmarshalling toml")
		}
	}
	return newConfig
}

func (config *Config) SetDefaults() {
	config.Receiver.DestDir = ExpandPath("~/Downloads")
	config.Receiver.WorkDir = ExpandPath("~/.opensend")
	config.Receiver.SkipZeroconf = false
	config.Sender.WorkDir = ExpandPath("~/.opensend")
	config.Targets = map[string]map[string]string{}
}
