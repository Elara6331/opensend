package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"
)

func ExpandPath(s string) string {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal().Err(err).Msg("Error getting home directory")
	}
	// Expand any environment variables in string
	expandedString := os.ExpandEnv(s)
	// If string starts with ~
	if strings.HasPrefix(expandedString, "~") {
		// Replace ~ with user's home directory
		expandedString = strings.Replace(expandedString, "~", homeDir, 1)
	}
	// Clean file path
	expandedString = filepath.Clean(expandedString)
	// Return expanded string
	return expandedString
}
