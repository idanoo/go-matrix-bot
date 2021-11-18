package gomatrixbot

import (
	"fmt"
	"strings"

	"maunium.net/go/mautrix/crypto"
)

// Simple crypto.Logger implementation that just prints to stdout.
type logger struct{}

var _ crypto.Logger = &logger{}

func (f logger) Error(message string, args ...interface{}) {
	fmt.Printf("[ERROR] "+message+"\n", args...)
}

func (f logger) Warn(message string, args ...interface{}) {
	fmt.Printf("[WARN] "+message+"\n", args...)
}

func (f logger) Debug(message string, args ...interface{}) {
	fmt.Printf("[DEBUG] "+message+"\n", args...)
}

func (f logger) Trace(message string, args ...interface{}) {
	if strings.HasPrefix(message, "Got membership state event") {
		return
	}
	// fmt.Printf("[TRACE] "+message+"\n", args...)
}
