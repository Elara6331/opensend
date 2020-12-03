package main

import (
	"github.com/rs/zerolog"
	"os"
)

type FatalHook struct {}

func (hook FatalHook) Run(_ *zerolog.Event, level zerolog.Level, _ string) {
	// If log event is fatal
	if level == zerolog.FatalLevel {
		// Attempt removal of opensend directory
		_ = os.RemoveAll(opensendDir)
	}
}
