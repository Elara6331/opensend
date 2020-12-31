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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal().Err(err).Msg("Error getting home directory")
	}
	expandedString := os.ExpandEnv(s)
	if strings.HasPrefix(expandedString, "~") {
		expandedString = strings.Replace(expandedString, "~", homeDir, 1)
	}
	expandedString = filepath.Clean(expandedString)
	return expandedString
}
