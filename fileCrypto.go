package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Encrypt given file using the shared key
func EncryptFile(filePath string, newFilePath string, sharedKey string) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
	// Read data from file
	data, err := ioutil.ReadFile(filePath)
	if err != nil { log.Fatal().Err(err).Msg("Error reading file") }
	// Create md5 hash of password in order to make it the required size
	md5Hash := md5.New()
	md5Hash.Write([]byte(sharedKey))
	// Encode md5 hash bytes into hexadecimal
	hashedKey := hex.EncodeToString(md5Hash.Sum(nil))
	// Create new AES cipher
	block, err := aes.NewCipher([]byte(hashedKey))
	if err != nil { log.Fatal().Err(err).Msg("Error creating AES cipher") }
	// Create GCM for AES cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil { log.Fatal().Err(err).Msg("Error creating GCM") }
	// Make byte slice for nonce
	nonce := make([]byte, gcm.NonceSize())
	// Read random bytes into nonce slice
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil { log.Fatal().Err(err).Msg("Error creating nonce") }
	// Encrypt data
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	// Create new file
	newFile, err := os.Create(newFilePath)
	if err != nil { log.Fatal().Err(err).Msg("Error creating file") }
	// Defer file close
	defer newFile.Close()
	// Write ciphertext to new file
	bytesWritten, err := newFile.Write(ciphertext)
	if err != nil { log.Fatal().Err(err).Msg("Error writing to file") }
	// Log bytes written and to which file
	log.Info().Str("file", filepath.Base(newFilePath)).Msg("Wrote " + strconv.Itoa(bytesWritten) + " bytes")
}

// Decrypt given file using the shared key
func DecryptFile(filePath string, newFilePath string, sharedKey string) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
	// Read data from file
	data, err := ioutil.ReadFile(filePath)
	if err != nil { log.Fatal().Err(err).Msg("Error reading file") }
	// Create md5 hash of password in order to make it the required size
	md5Hash := md5.New()
	md5Hash.Write([]byte(sharedKey))
	hashedKey := hex.EncodeToString(md5Hash.Sum(nil))
	// Create new AES cipher
	block, _ := aes.NewCipher([]byte(hashedKey))
	// Create GCM for AES cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil { log.Fatal().Err(err).Msg("Error creating GCM") }
	// Get standard GCM nonce size
	nonceSize := gcm.NonceSize()
	// Get nonce and ciphertext from data
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	// Decrypt data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil { log.Fatal().Err(err).Msg("Error decrypting data") }
	// Create new file
	newFile, err := os.Create(newFilePath)
	if err != nil { log.Fatal().Err(err).Msg("Error creating file") }
	// Defer file close
	defer newFile.Close()
	// Write ciphertext to new file
	bytesWritten, err := newFile.Write(plaintext)
	if err != nil { log.Fatal().Err(err).Msg("Error writing to file") }
	// Log bytes written and to which file
	log.Info().Str("file", filepath.Base(newFilePath)).Msg("Wrote " + strconv.Itoa(bytesWritten) + " bytes")
}

// Encrypt files in given directory using shared key
func EncryptFiles(dir string, sharedKey string) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
	// Walk given directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// If error reading, return err
		if err != nil { return err }
		// If file is not a directory and is not the key
		if !info.IsDir() && !strings.Contains(path, "key.aes"){
			// Encrypt the file using shared key, appending .enc
			EncryptFile(path, path + ".enc", sharedKey)
			// Remove unencrypted file
			err := os.Remove(path)
			if err != nil { return err }
		}
		// Return nil if no error occurs
		return nil
	})
	if err != nil { log.Fatal().Err(err).Msg("Error encrypting files") }
}

// Decrypt files in given directory using shared key
func DecryptFiles(dir string, sharedKey string) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
	// Walk given directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// If error reading, return err
		if err != nil { return err }
		// If file is not a directory and is encrypted
		if !info.IsDir() && strings.Contains(path, ".enc") {
			// Decrypt the file using the shared key, removing .enc
			DecryptFile(path, strings.TrimSuffix(path, ".enc"), sharedKey)
		}
		// Return nil if no errors occurred
		return nil
	})
	if err != nil { log.Fatal().Err(err).Msg("Error decrypting files") }
}