package main

import (
	"encoding/json"
	"github.com/pkg/browser"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

// Create config type to store action type and data
type Config struct {
	ActionType string
	ActionData string
}

// Instantiate and return a new Config struct
func NewConfig(actionType string, actionData string) *Config {
	return &Config{ActionType: actionType, ActionData: actionData}
}

// Create config file
func (config *Config) CreateFile(dir string) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// Create config file at given directory
	configFile, err := os.Create(dir + "/config.json")
	if err != nil { log.Fatal().Err(err).Msg("Error creating config file") }
	// Close config file at the end of this function
	defer configFile.Close()
	// Marshal given Config struct into a []byte
	jsonData, err := json.Marshal(config)
	if err != nil { log.Fatal().Err(err).Msg("Error encoding JSON") }
	// Write []byte to previously created config file
	bytesWritten, err := configFile.Write(jsonData)
	if err != nil { log.Fatal().Err(err).Msg("Error writing JSON to file") }
	// Log bytes written
	log.Info().Str("file", "config.json").Msg("Wrote " + strconv.Itoa(bytesWritten) + " bytes")
}

// Collect all required files into given directory
func (config *Config) CollectFiles(dir string) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// If action type is file
	if config.ActionType == "file" {
		// Open file path in config.ActionData
		src, err := os.Open(config.ActionData)
		if err != nil { log.Fatal().Err(err).Msg("Error opening file from config") }
		// Close source file at the end of this function
		defer src.Close()
		// Create new file with the same name at given directory
		dst, err := os.Create(dir + "/" + filepath.Base(config.ActionData))
		if err != nil { log.Fatal().Err(err).Msg("Error creating file") }
		// Close new file at the end of this function
		defer dst.Close()
		// Copy data from source file to destination file
		_, err = io.Copy(dst, src)
		if err != nil { log.Fatal().Err(err).Msg("Error copying data to file") }
		// Replace file path in config.ActionData with file name
		config.ActionData = filepath.Base(config.ActionData)
	}
}

// Read config file at given file path
func (config *Config) ReadFile(filePath string) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// Read file at filePath
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil { log.Fatal().Err(err).Msg("Error reading config file") }
	// Unmarshal data from JSON into config struct
	err = json.Unmarshal(fileData, config)
	if err != nil { log.Fatal().Err(err).Msg("Error decoding JSON") }
}

// Execute action specified in config
func (config *Config) ExecuteAction(srcDir string) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// If action is file
	if config.ActionType == "file" {
		// Open file from config at given directory
		src, err := os.Open(srcDir + "/" + config.ActionData)
		if err != nil { log.Fatal().Err(err).Msg("Error reading file from config") }
		// Close source file at the end of this function
		defer src.Close()
		// Get user's home directory
		homeDir, err := os.UserHomeDir()
		if err != nil { log.Fatal().Err(err).Msg("Error getting home directory") }
		// Create file in user's Downloads directory
		dst, err := os.Create(homeDir + "/Downloads/" + config.ActionData)
		if err != nil { log.Fatal().Err(err).Msg("Error creating file") }
		// Close destination file at the end of this function
		defer dst.Close()
		// Copy data from source file to destination file
		_, err = io.Copy(dst, src)
		if err != nil { log.Fatal().Err(err).Msg("Error copying data to file") }
	// If action is url
	} else if config.ActionType == "url" {
		// Attempt to open URL in browser
		err := browser.OpenURL(config.ActionData)
		if err != nil { log.Fatal().Err(err).Msg("Error opening browser") }
	// Catchall
	} else {
		// Log unknown action type
		log.Fatal().Msg("Unknown action type " + config.ActionType)
	}
}