package slogplus

import (
	"io"
	"log/slog"
	"os"
)

// NewLogger creates a logger using Handler.
func NewLogger(out io.Writer, opts *Options) *slog.Logger {
	return slog.New(New(out, opts))
}

// Setup sets the global default logger.
func Setup(out io.Writer, opts *Options) {
	slog.SetDefault(NewLogger(out, opts))
}

// SetupDefault configures the global logger with default options on stdout.
func SetupDefault() {
	Setup(os.Stdout, nil)
}

// SetupProduction configures the global logger for production on stdout.
func SetupProduction() {
	SetupProductionTo(os.Stdout)
}

// SetupProductionTo configures the global logger for production on out.
func SetupProductionTo(out io.Writer) {
	Setup(out, ProductionOptions())
}

// SetupDevelopment configures the global logger for development on stdout.
func SetupDevelopment() {
	SetupDevelopmentTo(os.Stdout)
}

// SetupDevelopmentTo configures the global logger for development on out.
func SetupDevelopmentTo(out io.Writer) {
	Setup(out, DevelopmentOptions())
}

// Preset groups common environment options.
type Preset struct {
	Production  *Options
	Development *Options
	Test        *Options
}

// DefaultPreset contains the built-in option presets.
var DefaultPreset = Preset{
	Production:  ProductionOptions(),
	Development: DevelopmentOptions(),
	Test:        TestOptions(),
}

// ProductionOptions returns options suitable for production logs.
func ProductionOptions() *Options {
	return &Options{
		Level:      slog.LevelInfo,
		AddSource:  false,
		TimeFormat: "2006/01/02 15:04:05",
	}
}

// DevelopmentOptions returns options suitable for local development logs.
func DevelopmentOptions() *Options {
	return &Options{
		Level:      slog.LevelDebug,
		AddSource:  true,
		TimeFormat: "2006/01/02 15:04:05",
	}
}

// TestOptions returns compact options for tests.
func TestOptions() *Options {
	return &Options{
		Level:      slog.LevelDebug,
		AddSource:  false,
		TimeFormat: "15:04:05",
	}
}

// LevelVar aliases slog.LevelVar for dynamic runtime level changes.
type LevelVar = slog.LevelVar

// NewLevelVar creates a new dynamic level holder.
func NewLevelVar(level slog.Level) *LevelVar {
	v := new(LevelVar)
	v.Set(level)
	return v
}
