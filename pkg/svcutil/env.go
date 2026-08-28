package svcutil

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// EnvInt returns os.Getenv(key) as int, or def if unset or invalid.
func EnvInt(key string, def int) int {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

// EnvInt64 returns os.Getenv(key) as int64, or def if unset or invalid.
func EnvInt64(key string, def int64) int64 {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return def
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return n
}

// EnvDuration returns os.Getenv(key) as time.Duration, or def if unset or invalid.
func EnvDuration(key string, def time.Duration) time.Duration {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return def
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return def
	}
	return d
}

// EnvBool returns os.Getenv(key) as bool, or def if unset or invalid.
// It accepts 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False.
func EnvBool(key string, def bool) bool {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return def
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return def
	}
	return b
}
