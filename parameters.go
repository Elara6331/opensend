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
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/mholt/archiver/v3"
	"github.com/pkg/browser"
	"github.com/rs/zerolog/log"
	"github.com/vmihailenco/msgpack/v5"
)

// Create config type to store action type and data
type Parameters struct {
	ActionType string
	ActionData string
}

// Instantiate and return a new Config struct
func NewParameters(actionType string, actionData string) *Parameters {
	return &Parameters{ActionType: actionType, ActionData: actionData}
}

func (parameters *Parameters) Validate() {
	if parameters.ActionType == "url" {
		// Parse URL in parameters
		urlParser, err := url.Parse(parameters.ActionData)
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
}

// Create config file
func (parameters *Parameters) CreateFile(dir string) {
	// Use ConsoleWriter logger
	// Create parameters file at given directory
	configFile, err := os.Create(dir + "/parameters.msgpack")
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating parameters file")
	}
	// Close parameters file at the end of this function
	defer configFile.Close()
	// Marshal given Parameters struct into a []byte
	MessagePackData, err := msgpack.Marshal(parameters)
	if err != nil {
		log.Fatal().Err(err).Msg("Error encoding MessagePack")
	}
	// Write []byte to previously created parameters file
	bytesWritten, err := configFile.Write(MessagePackData)
	if err != nil {
		log.Fatal().Err(err).Msg("Error writing MessagePack to file")
	}
	// Log bytes written
	log.Info().Str("file", "parameters.msgpack").Msg("Wrote " + strconv.Itoa(bytesWritten) + " bytes")
}

// Collect all required files into given directory
func (parameters *Parameters) CollectFiles(dir string) {
	// Use ConsoleWriter logger
	// If action type is file
	if parameters.ActionType == "file" {
		// Open file path in parameters.ActionData
		src, err := os.Open(parameters.ActionData)
		if err != nil {
			log.Fatal().Err(err).Msg("Error opening file from parameters")
		}
		// Close source file at the end of this function
		defer src.Close()
		// Create new file with the same name at given directory
		dst, err := os.Create(dir + "/" + filepath.Base(parameters.ActionData))
		if err != nil {
			log.Fatal().Err(err).Msg("Error creating file")
		}
		// Close new file at the end of this function
		defer dst.Close()
		// Copy data from source file to destination file
		_, err = io.Copy(dst, src)
		if err != nil {
			log.Fatal().Err(err).Msg("Error copying data to file")
		}
		// Replace file path in parameters.ActionData with file name
		parameters.ActionData = filepath.Base(parameters.ActionData)
	} else if parameters.ActionType == "dir" {
		err := archiver.Archive([]string{parameters.ActionData}, dir+"/"+filepath.Base(parameters.ActionData)+".tar")
		if err != nil {
			log.Fatal().Err(err).Msg("Error creating tar archive")
		}
		// Set parameters data to base path for receiver
		parameters.ActionData = filepath.Base(parameters.ActionData)
	}
}

// Read config file at given file path
func (parameters *Parameters) ReadFile(filePath string) {
	// Use ConsoleWriter logger
	// Read file at filePath
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading parameters file")
	}
	// Unmarshal data from MessagePack into parameters struct
	err = msgpack.Unmarshal(fileData, parameters)
	if err != nil {
		log.Fatal().Err(err).Msg("Error decoding MessagePack")
	}
}

// Execute action specified in config
func (parameters *Parameters) ExecuteAction(srcDir string, destDir string) {
	// Use ConsoleWriter logger
	// If action is file
	switch parameters.ActionType {
	case "file":
		// Open file from parameters at given directory
		src, err := os.Open(srcDir + "/" + parameters.ActionData)
		if err != nil {
			log.Fatal().Err(err).Msg("Error reading file from parameters")
		}
		// Close source file at the end of this function
		defer src.Close()
		// Create file in user's Downloads directory
		dst, err := os.Create(filepath.Clean(destDir) + "/" + parameters.ActionData)
		if err != nil {
			log.Fatal().Err(err).Msg("Error creating file")
		}
		// Close destination file at the end of this function
		defer dst.Close()
		// Copy data from source file to destination file
		_, err = io.Copy(dst, src)
		if err != nil {
			log.Fatal().Err(err).Msg("Error copying data to file")
		}
		// If action is url
	case "url":
		// Parse received URL
		urlParser, err := url.Parse(parameters.ActionData)
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
		err = browser.OpenURL(parameters.ActionData)
		if err != nil {
			log.Fatal().Err(err).Msg("Error opening browser")
		}
		// If action is dir
	case "dir":
		// Set destination directory to ~/Downloads/{dir name}
		dstDir := filepath.Dir(filepath.Clean(destDir) + "/" + parameters.ActionData)

		fmt.Println(dstDir)

		err := archiver.Unarchive(srcDir+"/"+parameters.ActionData+".tar", dstDir)
		if err != nil {
			log.Fatal().Err(err).Msg("Error extracting tar archive")
		}
		// Loop to recursively unarchive tar file
		// Catchall
	default:
		// Log unknown action type
		log.Fatal().Msg("Unknown action type " + parameters.ActionType)
	}
}
