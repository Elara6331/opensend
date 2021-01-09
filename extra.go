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
