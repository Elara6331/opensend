package main

import (
	"fmt"
	"github.com/rs/zerolog"
	"net"
	"os"
)

// Fatal hook to run in case of Fatal error
type FatalHook struct {}

// Run function on trigger
func (hook FatalHook) Run(_ *zerolog.Event, level zerolog.Level, _ string) {
	// If log event is fatal
	if level == zerolog.FatalLevel {
		// Attempt removal of opensend directory
		_ = os.RemoveAll(opensendDir)
	}
}

// TCP Fatal hook to run in case of Fatal error with open TCP connection
type TCPFatalHook struct {
	conn net.Conn
}

// Run function on trigger
func (hook TCPFatalHook) Run(_ *zerolog.Event, level zerolog.Level, _ string) {
	// If log event is fatal
	if level == zerolog.FatalLevel {
		// Send error to connection
		_, _ = fmt.Fprintln(hook.conn, "ERR;")
		// Close connection
		_ = hook.conn.Close()
		// Attempt removal of opensend directory
		_ = os.RemoveAll(opensendDir)
	}

}