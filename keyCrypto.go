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
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net"
	"os"
	"strings"
)

// Generate RSA keypair
func GenerateRSAKeypair() (*rsa.PrivateKey, *rsa.PublicKey) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
	// Generate private/public RSA keypair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal().Err(err).Msg("Error generating RSA keypair")
	}
	// Get public key
	publicKey := privateKey.PublicKey
	// Return keypair
	return privateKey, &publicKey
}

// Get public key from sender
func GetKey(connection net.Conn) []byte {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
	// Send key request to connection
	_, err := fmt.Fprintln(connection, "key;")
	if err != nil {
		log.Fatal().Err(err).Msg("Error sending key request")
	}
	// Read received message
	message, err := bufio.NewReader(connection).ReadString('\n')
	if err != nil {
		log.Fatal().Err(err).Msg("Error getting key")
	}
	// Process received message
	procMessage := strings.Split(strings.TrimSpace(message), ";")
	// If ok code returned
	if procMessage[0] == "OK" {
		// Decode received hex string into key
		key, err := hex.DecodeString(procMessage[1])
		if err != nil {
			log.Fatal().Err(err).Msg("Error reading key")
		}
		// Return key
		return key
		// Otherwise
	} else {
		// Fatally log
		if err != nil {
			log.Fatal().Msg("Server reported error")
		}
	}
	// Return nil if all else fails
	return nil
}

// Encrypt shared key with received public key
func EncryptKey(sharedKey string, recvPubKey *rsa.PublicKey) []byte {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
	// Encrypt shared key using RSA
	encryptedSharedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, recvPubKey, []byte(sharedKey), nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Error encrypting shared key")
	}
	// Return encrypted key
	return encryptedSharedKey
}

// Decrypt shared key using private RSA key
func DecryptKey(encryptedKey []byte, privateKey *rsa.PrivateKey) string {
	// Decrypt shared key using RSA
	decryptedKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedKey, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Error decrypting shared key")
	}
	// Get string of decrypted key
	sharedKey := string(decryptedKey)
	// Return shared key
	return sharedKey
}
