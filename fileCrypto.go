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
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"github.com/klauspost/compress/zstd"
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
func CompressAndEncryptFile(filePath string, newFilePath string, sharedKey string) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
	// Read data from file
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Error opening file")
	}
	// Create buffer for compressed data
	compressedBuffer := new(bytes.Buffer)
	// Create Zstd encoder
	zstdEncoder, err := zstd.NewWriter(compressedBuffer)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating Zstd encoder")
	}
	// Copy file data to Zstd encoder
	_, err = io.Copy(zstdEncoder, file)
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading file")
	}
	// Close Zstd encoder
	zstdEncoder.Close()
	// Read compressed data into data variable
	data, err := ioutil.ReadAll(compressedBuffer)
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading compressed buffer")
	}
	// Create md5 hash of password in order to make it the required size
	md5Hash := md5.New()
	md5Hash.Write([]byte(sharedKey))
	// Encode md5 hash bytes into hexadecimal
	hashedKey := hex.EncodeToString(md5Hash.Sum(nil))
	// Create new AES cipher
	block, err := aes.NewCipher([]byte(hashedKey))
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating AES cipher")
	}
	// Create GCM for AES cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating GCM")
	}
	// Make byte slice for nonce
	nonce := make([]byte, gcm.NonceSize())
	// Read random bytes into nonce slice
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating nonce")
	}
	// Encrypt data
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	// Create new file
	newFile, err := os.Create(newFilePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating file")
	}
	// Defer file close
	defer newFile.Close()
	// Write ciphertext to new file
	bytesWritten, err := newFile.Write(ciphertext)
	if err != nil {
		log.Fatal().Err(err).Msg("Error writing to file")
	}
	// Log bytes written and to which file
	log.Info().Str("file", filepath.Base(newFilePath)).Msg("Wrote " + strconv.Itoa(bytesWritten) + " bytes")
}

// Decrypt given file using the shared key
func DecryptAndDecompressFile(filePath string, newFilePath string, sharedKey string) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
	// Read data from file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading file")
	}
	// Create md5 hash of password in order to make it the required size
	md5Hash := md5.New()
	md5Hash.Write([]byte(sharedKey))
	hashedKey := hex.EncodeToString(md5Hash.Sum(nil))
	// Create new AES cipher
	block, _ := aes.NewCipher([]byte(hashedKey))
	// Create GCM for AES cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating GCM")
	}
	// Get standard GCM nonce size
	nonceSize := gcm.NonceSize()
	// Get nonce and ciphertext from data
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	// Decrypt data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Error decrypting data")
	}
	// Create new Zstd decoder
	zstdDecoder, err := zstd.NewReader(bytes.NewBuffer(plaintext))
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating Zstd decoder")
	}
	// Create new file
	newFile, err := os.Create(newFilePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating file")
	}
	// Close new file at the end of this function
	defer newFile.Close()
	// Write decompressed plaintext to new file
	bytesWritten, err := io.Copy(newFile, zstdDecoder)
	if err != nil {
		log.Fatal().Err(err).Msg("Error writing to file")
	}
	zstdDecoder.Close()
	// Log bytes written and to which file
	log.Info().Str("file", filepath.Base(newFilePath)).Msg("Wrote " + strconv.Itoa(int(bytesWritten)) + " bytes")
}

// Encrypt files in given directory using shared key
func EncryptFiles(dir string, sharedKey string) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
	// Walk given directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// If error reading, return err
		if err != nil {
			return err
		}
		// If file is not a directory and is not the key
		if !info.IsDir() && !strings.Contains(path, "key.aes") {
			// Compress and Encrypt the file using shared key, appending .zst.enc
			CompressAndEncryptFile(path, path+".zst.enc", sharedKey)
			// Remove unencrypted file
			err := os.Remove(path)
			if err != nil {
				return err
			}
		}
		// Return nil if no error occurs
		return nil
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Error encrypting files")
	}
}

// Decrypt files in given directory using shared key
func DecryptFiles(dir string, sharedKey string) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
	// Walk given directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// If error reading, return err
		if err != nil {
			return err
		}
		// If file is not a directory and is encrypted
		if !info.IsDir() && strings.Contains(path, ".enc") {
			// Decrypt and decompress the file using the shared key, removing .zst.enc
			DecryptAndDecompressFile(path, strings.TrimSuffix(path, ".zst.enc"), sharedKey)
		}
		// Return nil if no errors occurred
		return nil
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Error decrypting files")
	}
}
