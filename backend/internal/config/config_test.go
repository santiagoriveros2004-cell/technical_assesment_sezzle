package config_test

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/santiagoriveros2004-cell/technical_assesment_sezzle/backend/internal/config"
)

func environment(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func TestLoadDefaults(t *testing.T) {
	got, err := config.Load(environment(nil))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := config.Config{Port: 8080, LogLevel: slog.LevelInfo, ShutdownTimeout: 10 * time.Second}
	if got != want {
		t.Fatalf("Load() = %+v, want %+v", got, want)
	}
	if got.Addr() != ":8080" {
		t.Fatalf("Addr() = %q, want %q", got.Addr(), ":8080")
	}
}

func TestLoadOverrides(t *testing.T) {
	got, err := config.Load(environment(map[string]string{
		"PORT":             "9090",
		"LOG_LEVEL":        "debug",
		"SHUTDOWN_TIMEOUT": "1m30s",
	}))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := config.Config{Port: 9090, LogLevel: slog.LevelDebug, ShutdownTimeout: 90 * time.Second}
	if got != want {
		t.Fatalf("Load() = %+v, want %+v", got, want)
	}
}

func TestLoadLogLevels(t *testing.T) {
	for raw, want := range map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	} {
		got, err := config.Load(environment(map[string]string{"LOG_LEVEL": raw}))
		if err != nil {
			t.Fatalf("Load(LOG_LEVEL=%q) error = %v", raw, err)
		}
		if got.LogLevel != want {
			t.Fatalf("Load(LOG_LEVEL=%q).LogLevel = %v, want %v", raw, got.LogLevel, want)
		}
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{name: "port not a number", env: map[string]string{"PORT": "http"}, wantErr: "PORT"},
		{name: "port zero", env: map[string]string{"PORT": "0"}, wantErr: "PORT"},
		{name: "port above range", env: map[string]string{"PORT": "65536"}, wantErr: "PORT"},
		{name: "unknown log level", env: map[string]string{"LOG_LEVEL": "verbose"}, wantErr: "LOG_LEVEL"},
		{name: "timeout without unit", env: map[string]string{"SHUTDOWN_TIMEOUT": "10"}, wantErr: "SHUTDOWN_TIMEOUT"},
		{name: "negative timeout", env: map[string]string{"SHUTDOWN_TIMEOUT": "-1s"}, wantErr: "SHUTDOWN_TIMEOUT"},
		{name: "zero timeout", env: map[string]string{"SHUTDOWN_TIMEOUT": "0s"}, wantErr: "SHUTDOWN_TIMEOUT"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := config.Load(environment(tc.env))
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Load() error = %v, want error mentioning %s", err, tc.wantErr)
			}
		})
	}
}

func TestLoadReportsEveryInvalidValue(t *testing.T) {
	_, err := config.Load(environment(map[string]string{
		"PORT":             "x",
		"LOG_LEVEL":        "x",
		"SHUTDOWN_TIMEOUT": "x",
	}))
	if err == nil {
		t.Fatal("Load() error = nil")
	}
	for _, key := range []string{"PORT", "LOG_LEVEL", "SHUTDOWN_TIMEOUT"} {
		if !strings.Contains(err.Error(), key) {
			t.Fatalf("Load() error = %v, want it to mention %s", err, key)
		}
	}
}
