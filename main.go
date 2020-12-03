package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil { log.Fatal().Err(err).Msg("Error getting home directory") }
	// Define opensend directory as ~/.opensend
	opensendDir := homeDir + "/.opensend"

	// Create channel for signals
	sig := make(chan os.Signal, 1)
	// Send message on channel upon reception of SIGINT or SIGTERM
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	// Intercept signal
	go func() {
		select {
		// Wait for sig to be written
		case <-sig:
			// Warn user that a signal has been received and that opensend is shutting down
			log.Warn().Msg("Signal received. Shutting down.")
			// Remove opensend directory to avoid future conflicts
			_ = os.RemoveAll(opensendDir)
			// Exit with code 0
			os.Exit(0)
		}
	}()

	// Create -t flag for type
	actionType := flag.String("t", "","Type of data being sent")
	// Create -d flag for data
	actionData := flag.String("d", "", "Data to send")
	// Create -s flag for sending
	sendFlag := flag.Bool("s", false, "Send data")
	// Create -r flag for receiving
	recvFlag := flag.Bool("r", false, "Receive data")
	// Parse flags
	flag.Parse()

	// Create opensend dir ignoring errors
	_ = os.Mkdir(opensendDir, 0755)
	// If -s given
	if *sendFlag {
		// Create 32 byte buffer
		sharedKeyBytes := make([]byte, 32)
		// Read random bytes into buffer
		_, err := io.ReadFull(rand.Reader, sharedKeyBytes)
		if err != nil { log.Fatal().Err(err).Msg("Error generating random bytes") }
		// Encode random bytes to hexadecimal
		sharedKey := hex.EncodeToString(sharedKeyBytes)
		// Notify user a key has been created
		log.Info().Msg("Generated random shared key")
		// Notify user device discovery is beginning
		log.Info().Msg("Discovering opensend receivers")
		// Discover all _opensend._tcp.local. mDNS services
		discoveredReceivers, discoveredIPs := DiscoverReceivers()
		// Create reader for STDIN
		reader := bufio.NewReader(os.Stdin)
		// Print hostnames of each receiver
		for index, receiver := range discoveredReceivers {
			// Print hostname and index+1
			fmt.Println("[" + strconv.Itoa(index + 1) + "]", receiver)
		}
		// Prompt user for choice
		fmt.Print("Choose a receiver: ")
		choiceStr, _ := reader.ReadString('\n')
		// Convert input to int after trimming spaces
		choiceInt, err := strconv.Atoi(strings.TrimSpace(choiceStr))
		if err != nil { log.Fatal().Err(err).Msg("Error converting choice to int") }
		// Set choiceIndex to choiceInt-1 to allow for indexing
		choiceIndex := choiceInt - 1
		// Get IP of chosen receiver
		choiceIP := discoveredIPs[choiceIndex]
		// Notify user of key exchange
		log.Info().Msg("Performing key exchange")
		// Exchange RSA keys with receiver
		rawKey := SenderKeyExchange(choiceIP)
		// Inform user receiver key has been received
		log.Info().Msg("Receiver key received")
		// Encrypt shared key using RSA public key
		key := EncryptKey(sharedKey, rawKey)
		// Save encrypted key in opensend directory as savedKey.aesKey
		SaveEncryptedKey(key, opensendDir + "/savedKey.aesKey")
		// Instantiate Config object
		config := NewConfig(*actionType, *actionData)
		// Collect any files that may be required for transaction into opensend directory
		config.CollectFiles(opensendDir)
		// Create config file in opensend directory
		config.CreateFile(opensendDir)
		// Notify user file encryption is beginning
		log.Info().Msg("Encrypting files")
		// Encrypt all files in opensend directory using shared key
		EncryptFiles(opensendDir, sharedKey)
		// Notify user server has started
		log.Info().Msg("Server started on port 9898")
		// Send all files in opensend directory using an HTTP server on port 9898
		SendFiles(opensendDir)
	// If -r given
	} else if *recvFlag {
		// Register {hostname}._opensend._tcp.local. mDNS service and pass shutdown function
		zeroconfShutdown := RegisterService()
		// Shutdown zeroconf server at the end of main()
		defer zeroconfShutdown()
		// Notify user opensend is waiting for key exchange
		log.Info().Msg("Waiting for sender key exchange")
		// Generate keypair
		privateKey, publicKey := GenerateRSAKeypair()
		// Exchange keys with sender
		senderIP := ReceiverKeyExchange(publicKey)
		// Sleep 300ms to allow sender time to start HTTP server
		time.Sleep(300*time.Millisecond)
		// Notify user files are being received
		log.Info().Msg("Receiving files from server (This may take a while)")
		// Get files from sender and place them into the opensend directory
		RecvFiles(opensendDir, senderIP)
		// Get encrypted shared key from sender
		encryptedKey := GetKey(senderIP)
		// Send stop signal to sender's HTTP server
		SendSrvStopSignal(senderIP)
		// Decrypt shared key
		sharedKey := DecryptKey(encryptedKey, privateKey)
		// Notify user file decryption is beginning
		log.Info().Msg("Decrypting files")
		// Decrypt all files in opensend directory using shared key
		DecryptFiles(opensendDir, sharedKey)
		// Instantiate Config
		config := &Config{}
		// Read config file in opensend directory
		config.ReadFile(opensendDir + "/config.json")
		// Notify user that action is being executed
		log.Info().Msg("Executing JSON action")
		// Execute JSON action using files within opensend directory
		config.ExecuteAction(opensendDir)
	}
	// Remove opensend directory
	err = os.RemoveAll(opensendDir)
	if err != nil { log.Fatal().Err(err).Msg("Error remove opensend dir") }
}