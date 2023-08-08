package logger

import (
	"context"
	"github.com/sirupsen/logrus"
	"os"
)

type contextKey int

const loggerKey contextKey = iota

func NewLogger() *logrus.Logger {
	// Create a new logger instance
	log := logrus.New()

	// Set the log level
	log.SetLevel(logrus.DebugLevel)

	// Set the output to stdout
	log.SetOutput(os.Stdout)

	// Add a hook to log errors to Sentry (optional)
	// client, err := sentry.NewClient(...)
	// if err == nil {
	//     log.AddHook(&sentryhook{client})
	// }

	return log
}

func WithLogger(ctx context.Context, logger *logrus.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
