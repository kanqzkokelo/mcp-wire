package config

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// SetupSignalContext creates a context that cancels on SIGINT or SIGTERM.
func SetupSignalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}
