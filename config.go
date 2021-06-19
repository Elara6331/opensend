/*
   Copyright © 2021 Arsen Musayelyan

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog/log"
)

// Struct for unmarshaling of opensend TOML configs
type Config struct {
	Receiver ReceiverConfig
	Sender   SenderConfig
	Targets  map[string]Target
}

// Config section for receiver
type ReceiverConfig struct {
	DestDir      string `toml:"destinationDirectory"`
	SkipZeroconf bool
	WorkDir      string `toml:"workingDirectory"`
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
