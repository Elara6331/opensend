package main

import (
	"archive/tar"
	"encoding/json"
	"github.com/pkg/browser"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

func (config *Config) Validate()  {
	// Parse URL in config
	urlParser, err := url.Parse(config.ActionData)
	// If there was an error parsing
	if err != nil {
		// Alert user of invalid url
		log.Fatal().Err(err).Msg("Invalid URL")
	// If scheme is not detected
	} else if urlParser.Scheme == "" {
		// Alert user of invalid scheme
		log.Fatal().Msg("Invalid URL scheme")
	// If host is not detected
	} else if urlParser.Host == "" {
		// Alert user of invalid host
		log.Fatal().Msg("Invalid URL host")
	}
}

// Create config file
func (config *Config) CreateFile(dir string) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
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
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
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
	} else if config.ActionType == "dir" {
		// Create tar archive
		tarFile, err := os.Create(dir + "/" + filepath.Base(config.ActionData) + ".tar")
		if err != nil { log.Fatal().Err(err).Msg("Error creating file") }
		// Close tar file at the end of this function
		defer tarFile.Close()
		// Create writer for tar archive
		tarArchiver := tar.NewWriter(tarFile)
		// Close archiver at the end of this function
		defer tarArchiver.Close()
		// Walk given directory
		err = filepath.Walk(config.ActionData, func(path string, info os.FileInfo, err error) error {
			// Return if error walking
			if err != nil { return err }
			// Skip if file is not normal mode
			if !info.Mode().IsRegular() { return nil }
			// Create tar header for file
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil { return err }
			// Change header name to reflect decompressed filepath
			header.Name = strings.TrimPrefix(strings.ReplaceAll(path, config.ActionData, ""), string(filepath.Separator))
			// Write header to archive
			if err := tarArchiver.WriteHeader(header); err != nil { return err }
			// Open source file
			src, err := os.Open(path)
			if err != nil { return err }
			// Close source file at the end of this function
			defer src.Close()
			// Copy source bytes to tar archive
			if _, err := io.Copy(tarArchiver, src); err != nil { return err }
			// Return at the end of the function
			return nil
		})
		if err != nil { log.Fatal().Err(err).Msg("Error creating tar archive") }
		// Set config data to base path for receiver
		config.ActionData = filepath.Base(config.ActionData)
	}
}

// Read config file at given file path
func (config *Config) ReadFile(filePath string) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
	// Read file at filePath
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil { log.Fatal().Err(err).Msg("Error reading config file") }
	// Unmarshal data from JSON into config struct
	err = json.Unmarshal(fileData, config)
	if err != nil { log.Fatal().Err(err).Msg("Error decoding JSON") }
}

// Execute action specified in config
func (config *Config) ExecuteAction(srcDir string, destDir string) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
	// If action is file
	if config.ActionType == "file" {
		// Open file from config at given directory
		src, err := os.Open(srcDir + "/" + config.ActionData)
		if err != nil { log.Fatal().Err(err).Msg("Error reading file from config") }
		// Close source file at the end of this function
		defer src.Close()
		// Create file in user's Downloads directory
		dst, err := os.Create(filepath.Clean(destDir) + "/" + config.ActionData)
		if err != nil { log.Fatal().Err(err).Msg("Error creating file") }
		// Close destination file at the end of this function
		defer dst.Close()
		// Copy data from source file to destination file
		_, err = io.Copy(dst, src)
		if err != nil { log.Fatal().Err(err).Msg("Error copying data to file") }
	// If action is url
	} else if config.ActionType == "url" {
		// Parse received URL
		urlParser, err := url.Parse(config.ActionData)
		// If there was an error parsing
		if err != nil {
			// Alert user of invalid url
			log.Fatal().Err(err).Msg("Invalid URL")
		// If scheme is not detected
		} else if urlParser.Scheme == "" {
			// Alert user of invalid scheme
			log.Fatal().Msg("Invalid URL scheme")
		// If host is not detected
		} else if urlParser.Host == "" {
			// Alert user of invalid host
			log.Fatal().Msg("Invalid URL host")
		}
		// Attempt to open URL in browser
		err = browser.OpenURL(config.ActionData)
		if err != nil { log.Fatal().Err(err).Msg("Error opening browser") }
	// If action is dir
	} else if config.ActionType == "dir" {
		// Set destination directory to ~/Downloads/{dir name}
		dstDir := filepath.Clean(destDir) + "/" + config.ActionData
		// Try to create destination directory
		err := os.MkdirAll(dstDir, 0755)
		if err != nil { log.Fatal().Err(err).Msg("Error creating directory") }
		// Try to open tar archive file
		tarFile, err := os.Open(srcDir + "/" + config.ActionData + ".tar")
		if err != nil { log.Fatal().Err(err).Msg("Error opening tar archive") }
		// Close tar archive file at the end of this function
		defer tarFile.Close()
		// Create tar reader to unarchive tar archive
		tarUnarchiver := tar.NewReader(tarFile)
		// Loop to recursively unarchive tar file
		unarchiveLoop: for {
			// Jump to next header in tar archive
			header, err := tarUnarchiver.Next()
			switch {
			// If EOF
			case err == io.EOF:
				// break loop
				break unarchiveLoop
			case err != nil:
				log.Fatal().Err(err).Msg("Error unarchiving tar archive")
			// If nil header
			case header == nil:
				// Skip
				continue
			}
			// Set target path to header name in destination dir
			targetPath := filepath.Join(dstDir, header.Name)
			switch header.Typeflag {
			// If regular file
			case tar.TypeReg:
				// Try to create containing folder ignoring errors
				_ = os.MkdirAll(filepath.Dir(targetPath), 0755)
				// Create file with mode contained in header at target path
				dstFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
				if err != nil { log.Fatal().Err(err).Msg("Error creating file during unarchiving") }
				// Copy data from tar archive into file
				_, err = io.Copy(dstFile, tarUnarchiver)
				if err != nil { log.Fatal().Err(err).Msg("Error copying data to file") }
			}
		}
	// Catchall
	} else {
		// Log unknown action type
		log.Fatal().Msg("Unknown action type " + config.ActionType)
	}
}