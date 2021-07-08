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

package crypto

import (
	"crypto/rsa"
	"encoding/gob"
	"net"

	"github.com/rs/zerolog/log"
)

// Exchange keys with sender
func ReceiverKeyExchange(key *rsa.PublicKey) string {
	// Use ConsoleWriter logger
	// Create TCP listener on port 9797
	listener, err := net.Listen("tcp", ":9797")
	if err != nil {
		log.Fatal().Err(err).Msg("Error starting listener")
	}
	// Create string for sender address
	var senderAddr string
	for {
		// Accept connection on listener
		connection, err := listener.Accept()
		if err != nil {
			log.Fatal().Err(err).Msg("Error accepting connections")
		}
		// Get sender address and store it in senderAddr
		senderAddr = connection.RemoteAddr().String()
		// Create gob encoder with connection as io.Writer
		encoder := gob.NewEncoder(connection)
		// Encode key into connection
		err = encoder.Encode(key)
		if err != nil {
			log.Fatal().Err(err).Msg("Error encoding key")
		}
		return senderAddr
	}
}

// Exchange keys with receiver
func SenderKeyExchange(receiverIP string) *rsa.PublicKey {
	// Use ConsoleWriter logger
	// Connect to TCP socket on receiver IP port 9797
	connection, err := net.Dial("tcp", receiverIP+":9797")
	if err != nil {
		log.Fatal().Err(err).Msg("Error connecting to sender")
	}
	// Create gob decoder
	decoder := gob.NewDecoder(connection)
	// Instantiate rsa.PublicKey struct
	recvPubKey := &rsa.PublicKey{}
	// Decode key
	err = decoder.Decode(recvPubKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Error decoding key")
	}
	// Return received key
	return recvPubKey
}
