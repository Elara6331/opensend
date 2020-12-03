package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// Generate RSA keypair
func GenerateRSAKeypair() (*rsa.PrivateKey, *rsa.PublicKey) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// Generate private/public RSA keypair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil { log.Fatal().Err(err).Msg("Error generating RSA keypair") }
	// Get public key
	publicKey := privateKey.PublicKey
	// Return keypair
	return privateKey, &publicKey
}

// Get public key from sender
func GetKey(senderAddr string) []byte {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// Get server address by getting the IP without the port, prepending http:// and appending :9898
	serverAddr := "http://" + strings.Split(senderAddr, ":")[0] + ":9898"
	// GET /key on the sender's HTTP server
	response, err := http.Get(serverAddr + "/key")
	if err != nil { log.Fatal().Err(err).Msg("Error getting key") }
	// Close response body at the end of this function
	defer response.Body.Close()
	// If server responded with 200 OK
	if response.StatusCode == http.StatusOK {
		// Read response body into key
		key, err := ioutil.ReadAll(response.Body)
		if err != nil { log.Fatal().Err(err).Msg("Error reading HTTP response") }
		// Return key
		return key
	// Otherwise
	} else {
		// Fatally log status code
		if err != nil { log.Fatal().Int("code", response.StatusCode).Msg("HTTP Error Response Code Received") }
	}
	// Return nil if all else fails
	return nil
}

// Encrypt shared key with received public key
func EncryptKey(sharedKey string, recvPubKey *rsa.PublicKey) []byte {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	// Encrypt shared key using RSA
	encryptedSharedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, recvPubKey, []byte(sharedKey), nil)
	if err != nil { log.Fatal().Err(err).Msg("Error encrypting shared key") }
	// Return encrypted key
	return encryptedSharedKey
}

// Decrypt shared key using private RSA key
func DecryptKey(encryptedKey []byte, privateKey *rsa.PrivateKey) string {
	// Decrypt shared key using RSA
	decryptedKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedKey, nil)
	if err != nil { log.Fatal().Err(err).Msg("Error decrypting shared key") }
	// Get string of decrypted key
	sharedKey := string(decryptedKey)
	// Return shared key
	return sharedKey
}