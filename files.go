package main

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Save encrypted key to file
func SaveEncryptedKey(encryptedKey []byte, filePath string) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// Create file at given file path
	keyFile, err := os.Create(filePath)
	if err != nil { log.Fatal().Err(err).Msg("Error creating file") }
	// Close file at the end of this function
	defer keyFile.Close()
	// Write encrypted key to file
	bytesWritten, err := keyFile.Write(encryptedKey)
	if err != nil { log.Fatal().Err(err).Msg("Error writing key to file") }
	// Log bytes written
	log.Info().Str("file", filepath.Base(filePath)).Msg("Wrote " + strconv.Itoa(bytesWritten) + " bytes")
}

// Create HTTP server to transmit files
func SendFiles(dir string) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// Instantiate http.Server struct
	srv := &http.Server{}
	// Listen on all ipv4 addresses on port 9898
	listener, err := net.Listen("tcp4", ":9898")
	if err != nil { log.Fatal().Err(err).Msg("Error starting listener") }

	// If client connects to /:filePath
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		// Set file to first path components of URL, excluding first /
		file := req.URL.Path[1:]
		// Read file at specified location
		fileData, err := ioutil.ReadFile(dir + "/" + file)
		// If there was an error reading
		if err != nil {
			// Warn user of error
			log.Warn().Err(err).Msg("Error reading file")
		// Otherwise
		} else {
			// Inform user client has requested a file
			log.Info().Str("file", file).Msg("GET File")
		}
		// Write file to ResponseWriter
		_, err = fmt.Fprint(res, string(fileData))
		if err != nil { log.Fatal().Err(err).Msg("Error writing response") }
	})

	// If client connects to /index
	http.HandleFunc("/index", func(res http.ResponseWriter, req *http.Request) {
		// Inform user a client has requested the file index
		log.Info().Msg("GET Index")
		// Get directory listing
		dirListing, err := ioutil.ReadDir(dir)
		if err != nil { log.Fatal().Err(err).Msg("Error reading directory") }
		// Create new slice to house filenames for index
		var indexSlice []string
		// For each file in listing
		for _, file := range dirListing {
			// If the file is not the key
			if !strings.Contains(file.Name(), "savedKey.aesKey") {
				// Append the file path to indexSlice
				indexSlice = append(indexSlice, dir + "/" + file.Name())
			}
		}
		// Join index slice into string
		indexStr := strings.Join(indexSlice, ";")
		// Write index to ResponseWriter
		_, err = fmt.Fprint(res, indexStr)
		if err != nil { log.Fatal().Err(err).Msg("Error writing response") }
	})

	// If client connects to /key
	http.HandleFunc("/key", func(res http.ResponseWriter, req *http.Request) {
		// Inform user a client has requested the key
		log.Info().Msg("GET Key")
		// Read saved key
		key, err := ioutil.ReadFile(dir + "/savedKey.aesKey")
		if err != nil { log.Fatal().Err(err).Msg("Error reading key") }
		// Write saved key to ResponseWriter
		_, err = fmt.Fprint(res, string(key))
		if err != nil { log.Fatal().Err(err).Msg("Error writing response") }
	})

	// If client connects to /stop
	http.HandleFunc("/stop", func(res http.ResponseWriter, req *http.Request) {
		// Inform user a client has requested server shutdown
		log.Info().Msg("GET Stop")
		log.Info().Msg("Shutdown signal received")
		// Shutdown server and send to empty context
		err := srv.Shutdown(context.TODO())
		if err != nil { log.Fatal().Err(err).Msg("Error stopping server") }
	})

	// Start HTTP Server
	_ = srv.Serve(listener)
}

// Get files from sender
func RecvFiles(dir string, senderAddr string) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// Get server address by getting the IP without the port, prepending http:// and appending :9898
	serverAddr := "http://" + strings.Split(senderAddr, ":")[0] + ":9898"
	// GET /index on sender's HTTP server
	response, err := http.Get(serverAddr + "/index")
	if err != nil { log.Fatal().Err(err).Msg("Error getting index") }
	// Close response body at the end of this function
	defer response.Body.Close()
	// Create index slice for storage of file index
	var index []string
	// If server responded with 200 OK
	if response.StatusCode == http.StatusOK {
		// Read response body
		body, err := ioutil.ReadAll(response.Body)
		if err != nil { log.Fatal().Err(err).Msg("Error reading HTTP response") }
		// Get string from body
		bodyStr := string(body)
		// Split string to form index
		index = strings.Split(bodyStr, ";")
	}
	// For each file in the index
	for _, file := range index {
		// GET current file in index
		response, err := http.Get(serverAddr + "/" + filepath.Base(file))
		if err != nil { log.Fatal().Err(err).Msg("Error getting file") }
		// If server responded with 200 OK
		if response.StatusCode == http.StatusOK {
			// Create new file at index filepath
			newFile, err := os.Create(file)
			if err != nil { log.Fatal().Err(err).Msg("Error creating file") }
			// Copy response body to new file
			bytesWritten, err := io.Copy(newFile, response.Body)
			if err != nil { log.Fatal().Err(err).Msg("Error writing to file") }
			// Log bytes written
			log.Info().Str("file", filepath.Base(file)).Msg("Wrote " + strconv.Itoa(int(bytesWritten)) + " bytes")
			// Close new file
			newFile.Close()
		}
		// Close response body
		response.Body.Close()
	}
}

// Send stop signal to sender's HTTP server
func SendSrvStopSignal(senderAddr string) {
	// Get server address by getting the IP without the port, prepending http:// and appending :9898
	serverAddr := "http://" + strings.Split(senderAddr, ":")[0] + ":9898"
	// GET /stop on sender's HTTP servers ignoring any errors
	_, _ = http.Get(serverAddr + "/stop")
}