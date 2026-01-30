package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func Init(level string, pretty bool) {
	// Parse log level
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(lvl)

	if pretty {
		// Pretty console output for development
		Log = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Caller().Logger()
	} else {
		// JSON output for production
		Log = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}
}

// Convenience methods
func Debug() *zerolog.Event { return Log.Debug() }
func Info() *zerolog.Event  { return Log.Info() }
func Warn() *zerolog.Event  { return Log.Warn() }
func Error() *zerolog.Event { return Log.Error() }
func Fatal() *zerolog.Event { return Log.Fatal() }
