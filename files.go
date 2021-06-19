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
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

// Save encrypted key to file
func SaveEncryptedKey(encryptedKey []byte, filePath string) {
	// Use ConsoleWriter logger
	// Create file at given file path
	keyFile, err := os.Create(filePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating file")
	}
	// Close file at the end of this function
	defer keyFile.Close()
	// Write encrypted key to file
	bytesWritten, err := keyFile.Write(encryptedKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Error writing key to file")
	}
	// Log bytes written
	log.Info().Str("file", filepath.Base(filePath)).Msg("Wrote " + strconv.Itoa(bytesWritten) + " bytes")
}

// Create HTTP server to transmit files
func SendFiles(dir string) {
	// Use ConsoleWriter logger with normal FatalHook
	// Create TCP listener on port 9898
	listener, err := net.Listen("tcp", ":9898")
	if err != nil {
		log.Fatal().Err(err).Msg("Error starting listener")
	}

	http.HandleFunc("/key", func(res http.ResponseWriter, req *http.Request) {
		// Inform user client has requested key
		log.Info().Msg("Key requested")
		// Read saved key
		key, err := ioutil.ReadFile(dir + "/key.aes")
		if err != nil {
			log.Fatal().Err(err).Msg("Error reading key")
		}
		// Write saved key to ResponseWriter
		_, err = res.Write(key)
		if err != nil {
			log.Fatal().Err(err).Msg("Error writing response")
		}
	})

	http.HandleFunc("/index", func(res http.ResponseWriter, req *http.Request) {
		// Inform user a client has requested the file index
		log.Info().Msg("Index requested")
		// Get directory listing
		dirListing, err := ioutil.ReadDir(dir)
		if err != nil {
			log.Fatal().Err(err).Msg("Error reading directory")
		}
		// Create new slice to house filenames for index
		var indexSlice []string
		// For each file in listing
		for _, file := range dirListing {
			// If the file is not the key
			if !strings.Contains(file.Name(), "key.aes") {
				// Append the file path to indexSlice
				indexSlice = append(indexSlice, file.Name())
			}
		}
		// Join index slice into string
		indexStr := strings.Join(indexSlice, "|")
		// Write index to ResponseWriter
		_, err = res.Write([]byte(indexStr))
		if err != nil {
			log.Fatal().Err(err).Msg("Error writing response")
		}
	})

	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		log.Info().Str("file", filepath.Base(req.URL.Path)).Msg("File requested")
		http.FileServer(http.Dir(dir)).ServeHTTP(res, req)
	})

	http.HandleFunc("/stop", func(res http.ResponseWriter, req *http.Request) {
		log.Info().Msg("Stop signal received")
		res.WriteHeader(http.StatusOK)
		listener.Close()
	})

	http.Serve(listener, nil)
}

type Sender struct {
	RemoteAddr string
}

func (c *Sender) Get(endpoint string) (io.ReadCloser, int, error) {
	res, err := http.Get(c.RemoteAddr + endpoint)
	if err != nil {
		return nil, 0, err
	}
	return res.Body, res.StatusCode, nil
}

func NewSender(senderAddr string) *Sender {
	// Get server address by getting the IP without the port, and appending :9898
	host, _, _ := net.SplitHostPort(senderAddr)
	serverAddr := "http://" + net.JoinHostPort(host, "9898")
	return &Sender{RemoteAddr: serverAddr}
}

// Get files from sender
func RecvFiles(sender *Sender) {
	// Use ConsoleWriter logger
	indexReader, code, err := sender.Get("/index")
	if err != nil {
		log.Fatal().Err(err).Msg("Error getting index")
	}
	// If non-ok code returned, fatally log
	if code != http.StatusOK {
		log.Fatal().Err(err).Msg("Sender reported error")
	}
	indexBytes, err := ioutil.ReadAll(indexReader)
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading index from response")
	}
	// Get index from message
	index := strings.Split(strings.TrimSpace(string(indexBytes)), "|")
	for _, file := range index {
		// Read received message
		fileData, code, err := sender.Get("/" + file)
		if err != nil {
			log.Fatal().Err(err).Msg("Error getting file")
		}
		// If non-ok code returned
		if code != http.StatusOK {
			// fatally log
			log.Fatal().
				Int("status", code).
				Str("statusText", http.StatusText(code)).
				Err(err).
				Msg("Sender reported error")
			// Otherwise
		} else {
			// Create new file at index filepath
			newFile, err := os.Create(*workDir + "/" + file)
			if err != nil {
				log.Fatal().Err(err).Msg("Error creating file")
			}
			// Copy response body to new file
			bytesWritten, err := io.Copy(newFile, fileData)
			if err != nil {
				log.Fatal().Err(err).Msg("Error writing to file")
			}
			// Log bytes written
			log.Info().Str("file", filepath.Base(file)).Msg("Wrote " + strconv.Itoa(int(bytesWritten)) + " bytes")
			// Close new file
			newFile.Close()
		}
	}
}

// Send stop signal to sender
func SendSrvStopSignal(sender *Sender) {
	_, _, _ = sender.Get("/stop")
}
