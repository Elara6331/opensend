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
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
)

var workDir *string
var destDir *string

func main() {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})

	// Create --send-to flag to send to a specific IP
	sendTo := flag.String("send-to", "", "Use IP address of receiver instead of mDNS")
	// Create --dest-dir flag to save to a specified folder
	destDir = flag.String("dest-dir", "", "Destination directory for files or dirs sent over opensend")
	// Create --work-dir flag to perform operations in a specified directory
	workDir = flag.String("work-dir", "", "Working directory for opensend")
	// Create --config to select config file to use
	givenCfgPath := flag.String("config", "", "Opensend config to use")
	// Create --skip-mdns to skip service registration
	skipMdns := flag.Bool("skip-mdns", false, "Skip zeroconf service registration (use if mdns fails)")
	// Create -t flag for type
	actionType := flag.StringP("type", "t", "", "Type of data being sent")
	// Create -d flag for data
	actionData := flag.StringP("data", "d", "", "Data to send")
	// Create -s flag for sending
	sendFlag := flag.BoolP("send", "s", false, "Send data")
	// Create -r flag for receiving
	recvFlag := flag.BoolP("receive", "r", false, "Receive data")
	targetFlag := flag.StringP("target", "T", "", "Target as defined in opensend.toml")
	loopFlag := flag.BoolP("loop", "L", false, "Continuously wait for connections and handle them concurrently")
	// Parse flags
	flag.Parse()

	// Declare config variable
	var config *Config
	// If config flag not provided
	if *givenCfgPath == "" {
		// Get config path
		confPath := GetConfigPath()
		// Read config at path
		config = NewConfig(confPath)
	} else {
		// Otherwise, read config at provided path
		config = NewConfig(*givenCfgPath)
	}

	// If work directory flag not provided
	if *workDir == "" {
		// If send flag provided
		if *sendFlag {
			// Set work directory to sender as defined in config
			*workDir = ExpandPath(config.Sender.WorkDir)
		} else {
			// Otherwise set work directory to receiver as defined in config
			*workDir = ExpandPath(config.Receiver.WorkDir)
		}
	}

	// If destination directory flag not provided
	if *destDir == "" {
		// If receiver flag provided
		if *recvFlag {
			// Set destination directory to receiver as defined in config
			*destDir = ExpandPath(config.Receiver.DestDir)
		}
	}

	// If target flag provided
	if *targetFlag != "" {
		// Set IP to target's IP
		*sendTo = config.Targets[*targetFlag].IP
	}

	// Create channel for signals
	sig := make(chan os.Signal, 1)
	// Send message on channel upon reception of SIGINT or SIGTERM
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	// Intercept signal
	go func() {
		signal := <-sig
		// Warn user that a signal has been received and that opensend is shutting down
		log.Warn().Str("signal", signal.String()).Msg("Signal received. Shutting down.")
		// Remove opensend directory to avoid future conflicts
		_ = os.RemoveAll(*workDir)
		// Exit with code 0
		os.Exit(0)
	}()

	// Create opensend dir ignoring errors
	_ = os.Mkdir(*workDir, 0755)
	// If -s given
	if *sendFlag {
		if *actionType == "" || *actionData == "" {
			log.Fatal().Msg("Valid action type and data is required to send")
		}
		// Create 32 byte buffer
		sharedKeyBytes := make([]byte, 32)
		// Read random bytes into buffer
		_, err := io.ReadFull(rand.Reader, sharedKeyBytes)
		if err != nil {
			log.Fatal().Err(err).Msg("Error generating random bytes")
		}
		// Encode random bytes to hexadecimal
		sharedKey := hex.EncodeToString(sharedKeyBytes)
		// Notify user a key has been created
		log.Info().Msg("Generated random shared key")
		// Create variable to store chosen IP
		var choiceIP string
		// If IP is provided via --send-to
		if *sendTo != "" {
			// Notify user that provided IP is being used
			log.Info().Msg("IP provided. Skipping discovery.")
			// Set chosen IP to provided
			choiceIP = *sendTo
			// Otherwise
		} else {
			// Notify user device discovery is beginning
			log.Info().Msg("Discovering opensend receivers")
			// Discover all _opensend._tcp.local. mDNS services
			discoveredReceivers, discoveredIPs := DiscoverReceivers()
			// Create reader for STDIN
			reader := bufio.NewReader(os.Stdin)
			// Print hostnames of each receiver
			for index, receiver := range discoveredReceivers {
				// Print hostname and index+1
				fmt.Println("["+strconv.Itoa(index+1)+"]", receiver)
			}
			// Prompt user for choice
			fmt.Print("Choose a receiver: ")
			choiceStr, _ := reader.ReadString('\n')
			// Convert input to int after trimming spaces
			choiceInt, err := strconv.Atoi(strings.TrimSpace(choiceStr))
			if err != nil {
				log.Fatal().Err(err).Msg("Error converting choice to int")
			}
			// Set choiceIndex to choiceInt-1 to allow for indexing
			choiceIndex := choiceInt - 1
			// Get IP of chosen receiver
			choiceIP = discoveredIPs[choiceIndex]
		}
		// Instantiate Config object
		parameters := NewParameters(*actionType, *actionData)
		// Validate data in config struct
		parameters.Validate()
		// Collect any files that may be required for transaction into opensend directory
		parameters.CollectFiles(*workDir)
		// Create config file in opensend directory
		parameters.CreateFile(*workDir)
		// Notify user of key exchange
		log.Info().Msg("Performing key exchange")
		// Exchange RSA keys with receiver
		rawKey := SenderKeyExchange(choiceIP)
		// Inform user receiver key has been received
		log.Info().Msg("Receiver key received")
		// Encrypt shared key using RSA public key
		key := EncryptKey(sharedKey, rawKey)
		// Save encrypted key in opensend directory as key.aes
		SaveEncryptedKey(key, *workDir+"/key.aes")
		// Notify user file encryption is beginning
		log.Info().Msg("Encrypting files")
		// Encrypt all files in opensend directory using shared key
		EncryptFiles(*workDir, sharedKey)
		// Notify user server has started
		log.Info().Msg("Server started on port 9898")
		// Send all files in opensend directory using an HTTP server on port 9898
		SendFiles(*workDir)
	} else if *recvFlag && *loopFlag {
		// Declare zeroconf shutdown variable
		var zeroconfShutdown func()
		for {
			// Create opensend dir ignoring errors
			_ = os.Mkdir(*workDir, 0755)
			// If --skip-mdns is not given
			if !*skipMdns {
				// Register {hostname}._opensend._tcp.local. mDNS service and pass shutdown function
				zeroconfShutdown = RegisterService()
			}
			// Notify user keypair is being generated
			log.Info().Msg("Generating RSA keypair")
			// Generate keypair
			privateKey, publicKey := GenerateRSAKeypair()
			// Notify user opensend is waiting for key exchange
			log.Info().Msg("Waiting for sender key exchange")
			// Exchange keys with sender
			senderIP := ReceiverKeyExchange(publicKey)
			// If --skip-mdns is not given
			if !*skipMdns {
				// Shutdown zeroconf service as connection will be unavailable during transfer
				zeroconfShutdown()
			}
			// Sleep 300ms to allow sender time to start HTTP server
			time.Sleep(300 * time.Millisecond)
			// Notify user files are being received
			log.Info().Msg("Receiving files from server (This may take a while)")
			// Connect to sender's TCP socket
			connection := NewSender(senderIP)
			// Get files from sender and place them into the opensend directory
			RecvFiles(connection)
			// Get encrypted shared key from sender
			encryptedKey := GetKey(connection)
			// Send stop signal to sender's HTTP server
			SendSrvStopSignal(connection)
			// Decrypt shared key
			sharedKey := DecryptKey(encryptedKey, privateKey)
			// Notify user file decryption is beginning
			log.Info().Msg("Decrypting files")
			// Decrypt all files in opensend directory using shared key
			DecryptFiles(*workDir, sharedKey)
			// Instantiate Config
			parameters := &Parameters{}
			// Read config file in opensend directory
			parameters.ReadFile(*workDir + "/parameters.msgpack")
			// Notify user that action is being executed
			log.Info().Msg("Executing action")
			// Execute MessagePack action using files within opensend directory
			parameters.ExecuteAction(*workDir, *destDir)
			// Remove opensend directory
			err := os.RemoveAll(*workDir)
			if err != nil {
				log.Fatal().Err(err).Msg("Error removing opensend directory")
			}
		}
	} else if *recvFlag {
		// If --skip-mdns is not given
		if !*skipMdns {
			// Register {hostname}._opensend._tcp.local. mDNS service and pass shutdown function
			zeroconfShutdown := RegisterService()
			// Shutdown zeroconf server at the end of main()
			defer zeroconfShutdown()
		}
		// Notify user keypair is being generated
		log.Info().Msg("Generating RSA keypair")
		// Generate keypair
		privateKey, publicKey := GenerateRSAKeypair()
		// Notify user opensend is waiting for key exchange
		log.Info().Msg("Waiting for sender key exchange")
		// Exchange keys with sender
		senderIP := ReceiverKeyExchange(publicKey)
		// Sleep 300ms to allow sender time to start HTTP server
		time.Sleep(300 * time.Millisecond)
		// Notify user files are being received
		log.Info().Msg("Receiving files from server (This may take a while)")
		// Connect to sender's TCP socket
		sender := NewSender(senderIP)
		// Get files from sender and place them into the opensend directory
		RecvFiles(sender)
		// Get encrypted shared key from sender
		encryptedKey := GetKey(sender)
		// Send stop signal to sender's HTTP server
		SendSrvStopSignal(sender)
		// Decrypt shared key
		sharedKey := DecryptKey(encryptedKey, privateKey)
		// Notify user file decryption is beginning
		log.Info().Msg("Decrypting files")
		// Decrypt all files in opensend directory using shared key
		DecryptFiles(*workDir, sharedKey)
		// Instantiate Config
		parameters := &Parameters{}
		// Read config file in opensend directory
		parameters.ReadFile(*workDir + "/parameters.msgpack")
		// Notify user that action is being executed
		log.Info().Msg("Executing Action")
		// Execute MessagePack action using files within opensend directory
		parameters.ExecuteAction(*workDir, *destDir)
	} else {
		flag.Usage()
		log.Fatal().Msg("You must choose sender or receiver mode using -s or -r")
	}
	// Remove opensend directory
	err := os.RemoveAll(*workDir)
	if err != nil {
		log.Fatal().Err(err).Msg("Error removing opensend directory")
	}
}
