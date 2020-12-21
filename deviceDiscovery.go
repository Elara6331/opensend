package main

import (
	"context"
	"github.com/grandcat/zeroconf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

// Discover opensend receivers on the network
func DiscoverReceivers() ([]string, []string) {
	// Use ConsoleWriter logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Hook(FatalHook{})
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
