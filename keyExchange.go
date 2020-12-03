package main

import (
	"crypto/rsa"
	"encoding/gob"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net"
	"os"
)

// Exchange keys with sender
func ReceiverKeyExchange(key *rsa.PublicKey) string {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// Create TCP listener on port 9797
	listener, err := net.Listen("tcp", ":9797")
	if err != nil { log.Fatal().Err(err).Msg("Error starting listener") }
	// Create string for sender address
	var senderAddr string
	// Create channel to send break signal
	breakChannel := make(chan bool)
	for {
		// Accept connection on listener
		connection, err := listener.Accept()
		// Get sender address and store it in senderAddr
		senderAddr = connection.RemoteAddr().String()
		if err != nil { log.Fatal().Err(err).Msg("Error accepting connections") }
		// Concurrently handle connection
		go func(conn net.Conn) {
			// Create gob encoder with connection as io.Writer
			encoder := gob.NewEncoder(conn)
			// Encode key into connection
			err := encoder.Encode(key)
			if err != nil { log.Fatal().Err(err).Msg("Error encoding key") }
			// Send signal to breakChannel
			breakChannel <- true
		}(connection)
		// Wait for break signal
		select {
		// When break signal arrives
		case _ = <-breakChannel:
			// Return sender address
			return senderAddr
		}
	}
}

// Exchange keys with receiver
func SenderKeyExchange(receiverIP string) *rsa.PublicKey {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// Connect to TCP socket on receiver IP port 9797
	connection, err := net.Dial("tcp", receiverIP + ":9797")
	if err != nil { log.Fatal().Err(err).Msg("Error connecting to sender") }
	// Create gob decoder
	decoder := gob.NewDecoder(connection)
	// Instantiate rsa.PublicKey struct
	recvPubKey := &rsa.PublicKey{}
	// Decode key
	err = decoder.Decode(recvPubKey)
	if err != nil { log.Fatal().Err(err).Msg("Error decoding key") }
	// Return received key
	return recvPubKey
}
