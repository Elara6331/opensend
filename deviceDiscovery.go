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
	"context"
	"os"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/rs/zerolog/log"
)

// Discover opensend receivers on the network
func DiscoverReceivers() ([]string, []string) {
	// Use ConsoleWriter logger
	// Create zeroconf resolver
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating zeroconf resolver")
	}
	// Create channel for zeroconf entries
	entries := make(chan *zeroconf.ServiceEntry)
	// Create slice to store hostnames of discovered receivers
	var discoveredReceivers []string
	// Create slice to store IPs of discovered receivers
	var discoveredReceiverIPs []string
	// Concurrently run mDNS query
	go func(results <-chan *zeroconf.ServiceEntry) {
		// For each entry
		for entry := range results {
			// Append hostname to discoveredReceivers
			discoveredReceivers = append(discoveredReceivers, entry.HostName)
			// Append IP to discoveredReceiverIPs
			discoveredReceiverIPs = append(discoveredReceiverIPs, entry.AddrIPv4[0].String())
		}
	}(entries)

	// Create context with 4 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	// Cancel context at the end of this function
	defer cancel()
	// Browse for mDNS entries
	err = resolver.Browse(ctx, "_opensend._tcp", "local.", entries)
	if err != nil {
		log.Fatal().Err(err).Msg("Error browsing zeroconf services")
	}

	// Send Done signal to context
	<-ctx.Done()
	// Return discovered receiver slices
	return discoveredReceivers, discoveredReceiverIPs
}

// Register opensend zeroconf service on the network
func RegisterService() func() {
	// Get computer hostname
	hostname, _ := os.Hostname()
	// Register zeroconf service {hostname}._opensend._tcp.local.
	server, err := zeroconf.Register(hostname, "_opensend._tcp", "local.", 9797, []string{"txtv=0", "lo=1", "la=2"}, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Error registering zeroconf service")
	}
	// Return server.Shutdown() function to allow for shutdown in main()
	return server.Shutdown
}
