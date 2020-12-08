package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Save encrypted key to file
func SaveEncryptedKey(encryptedKey []byte, filePath string) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
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
	// Use ConsoleWriter logger with normal FatalHook
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
	// Create TCP listener on port 9898
	listener, err := net.Listen("tcp", ":9898")
	if err != nil { log.Fatal().Err(err).Msg("Error starting listener") }
	// Accept connection on listener
	connection, err := listener.Accept()
	if err != nil { log.Fatal().Err(err).Msg("Error accepting connection") }
	// Close connection at the end of this function
	defer connection.Close()
	// Create for loop to listen for messages on connection
	connectionLoop: for {
		// Use ConsoleWriter logger with TCPFatalHook
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(TCPFatalHook{conn: connection})
		// Attempt to read new message on connection
		data, err := bufio.NewReader(connection).ReadString('\n')
		// If no message detected, try again
		if err != nil && err.Error() == "EOF" { continue }
		// If non-EOF error, fatally log
		if err != nil { log.Fatal().Err(err).Msg("Error reading data") }
		// Process received data
		processedData := strings.Split(strings.TrimSpace(data), ";")
		// If processedData is empty, alert the user of invalid data
		if len(processedData) < 1 { log.Fatal().Str("data", data).Msg("Received data invalid") }
		switch processedData[0] {
		case "key":
			// Inform user client has requested key
			log.Info().Msg("Key requested")
			// Read saved key
			key, err := ioutil.ReadFile(dir + "/key.aes")
			if err != nil { log.Fatal().Err(err).Msg("Error reading key") }
			// Write saved key to ResponseWriter
			_, err = fmt.Fprintln(connection, "OK;" + hex.EncodeToString(key) + ";")
			if err != nil { log.Fatal().Err(err).Msg("Error writing response") }
		case "index":
			// Inform user a client has requested the file index
			log.Info().Msg("Index requested")
			// Get directory listing
			dirListing, err := ioutil.ReadDir(dir)
			if err != nil { log.Fatal().Err(err).Msg("Error reading directory") }
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
			_, err = fmt.Fprintln(connection, "OK;" + indexStr + ";")
			if err != nil { log.Fatal().Err(err).Msg("Error writing response") }
		case "file":
			// If processedData only has one entry
			if len(processedData) == 1 {
				// Warn user of unexpected end of line
				log.Warn().Err(errors.New("unexpected eol")).Msg("Invalid file request")
				// Send error to connection
				_, _ = fmt.Fprintln(connection, "ERR;")
				// Break out of switch
				break
			}
			// Set file to first path components of URL, excluding first /
			file := processedData[1]
			// Read file at specified location
			fileData, err := ioutil.ReadFile(dir + "/" + file)
			// If there was an error reading
			if err != nil {
				// Warn user of error
				log.Warn().Err(err).Msg("Error reading file")
				// Otherwise
			} else {
				// Inform user client has requested a file
				log.Info().Str("file", file).Msg("File requested")
			}
			// Write file as hex to connection
			_, err = fmt.Fprintln(connection, "OK;" + hex.EncodeToString(fileData) + ";")
			if err != nil { log.Fatal().Err(err).Msg("Error writing response") }
		case "stop":
			// Alert user that stop signal has been received
			log.Info().Msg("Received stop signal")
			// Print ok message to connection
			_, _ = fmt.Fprintln(connection, "OK;")
			// Break out of connectionLoop
			break connectionLoop
		}
	}
}

func ConnectToSender(senderAddr string) net.Conn {
	// Get server address by getting the IP without the port, and appending :9898
	serverAddr := strings.Split(senderAddr, ":")[0] + ":9898"
	// Create error variable
	var err error
	// Create connection variable
	var connection net.Conn
	// Until break
	for {
		// Try connecting to sender
		connection, err = net.Dial("tcp", serverAddr)
		// If connection refused
		if err != nil && strings.Contains(err.Error(), "connection refused") {
			// Continue loop (retry)
			continue
			// If error other than connection refused
		} else if err != nil {
			// Fatally log
			log.Fatal().Err(err).Msg("Error connecting to sender")
			// If no error
		} else {
			// Break out of loop
			break
		}
	}
	// Returned created connection
	return connection
}

// Get files from sender
func RecvFiles(connection net.Conn) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
	// Request index from sender
	_, err := fmt.Fprintln(connection, "index;")
	if err != nil { log.Fatal().Err(err).Msg("Error sending index request") }
	// Read received message
	message, err := bufio.NewReader(connection).ReadString('\n')
	if err != nil { log.Fatal().Err(err).Msg("Error getting index") }
	// Process received message
	procMessage := strings.Split(strings.TrimSpace(message), ";")
	// If non-ok code returned, fatally log
	if procMessage[0] != "OK" { log.Fatal().Err(err).Msg("Sender reported error") }
	// Get index from message
	index := strings.Split(strings.TrimSpace(procMessage[1]), "|")
	for _, file := range index {
		// Get current file in index
		_, err = fmt.Fprintln(connection, "file;" + file + ";")
		if err != nil { log.Fatal().Err(err).Msg("Error sending file request") }
		// Read received message
		message, err := bufio.NewReader(connection).ReadString('\n')
		if err != nil { log.Fatal().Err(err).Msg("Error getting file") }
		// Process received message
		procMessage := strings.Split(message, ";")
		// If non-ok code returned
		if procMessage[0] != "OK" {
			// fatally log
			log.Fatal().Err(err).Msg("Sender reported error")
		// Otherwise
		} else {
			// Create new file at index filepath
			newFile, err := os.Create(opensendDir + "/" + file)
			if err != nil { log.Fatal().Err(err).Msg("Error creating file") }
			// Decode file data from hex string
			fileData, err := hex.DecodeString(strings.TrimSpace(procMessage[1]))
			if err != nil { log.Fatal().Err(err).Msg("Error decoding hex") }
			// Copy response body to new file
			bytesWritten, err := io.Copy(newFile, bytes.NewBuffer(fileData))
			if err != nil { log.Fatal().Err(err).Msg("Error writing to file") }
			// Log bytes written
			log.Info().Str("file", filepath.Base(file)).Msg("Wrote " + strconv.Itoa(int(bytesWritten)) + " bytes")
			// Close new file
			newFile.Close()
		}
	}
}

// Send stop signal to sender
func SendSrvStopSignal(connection net.Conn) {
	// Send stop signal to connection
	_, _ = fmt.Fprintln(connection, "stop;")
	// Close connection
	_ = connection.Close()
}