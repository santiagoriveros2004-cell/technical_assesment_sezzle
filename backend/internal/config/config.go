package config

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"
)

type Config struct {
	Port            int
	LogLevel        slog.Level
	ShutdownTimeout time.Duration
}

var logLevels = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

func Load(getenv func(string) string) (Config, error) {
	port, portErr := parsePort(valueOrDefault(getenv("PORT"), "8080"))
	logLevel, logLevelErr := parseLogLevel(valueOrDefault(getenv("LOG_LEVEL"), "info"))
	shutdownTimeout, shutdownErr := parseShutdownTimeout(valueOrDefault(getenv("SHUTDOWN_TIMEOUT"), "10s"))

	if err := errors.Join(portErr, logLevelErr, shutdownErr); err != nil {
		return Config{}, err
	}
	return Config{Port: port, LogLevel: logLevel, ShutdownTimeout: shutdownTimeout}, nil
}

func (c Config) Addr() string {
	return ":" + strconv.Itoa(c.Port)
}

func parsePort(raw string) (int, error) {
	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("PORT must be an integer between 1 and 65535, got %q", raw)
	}
	return port, nil
}

func parseLogLevel(raw string) (slog.Level, error) {
	level, ok := logLevels[raw]
	if !ok {
		return 0, fmt.Errorf("LOG_LEVEL must be one of debug, info, warn, error, got %q", raw)
	}
	return level, nil
}

func parseShutdownTimeout(raw string) (time.Duration, error) {
	timeout, err := time.ParseDuration(raw)
	if err != nil || timeout <= 0 {
		return 0, fmt.Errorf("SHUTDOWN_TIMEOUT must be a positive duration such as 10s, got %q", raw)
	}
	return timeout, nil
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
