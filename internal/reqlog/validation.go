package reqlog

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var keyPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
var identifierListPattern = regexp.MustCompile(`^[a-zA-Z0-9\-\_\*,]+$`)

func validateDir(dir string, allowedDirs map[string]string) (string, error) {
	if dir == "" {
		return "", errors.New("directory is required")
	}
	for k, allowedDir := range allowedDirs {
		if dir == k {
			return allowedDir, nil
		}
	}
	return "", errors.New("directory not allowed")
}

func validateLimit(s string, defaultVal int, max int) int {
	limit, err := strconv.Atoi(s)
	if err != nil || limit < 0 || limit > max {
		return defaultVal
	}
	return limit
}

func validateSince(s string) (string, error) {
	s = strings.TrimSpace(s)

	if s == "" {
		return "", nil
	}

	// duration: 5m, 1h, 24h
	if _, err := time.ParseDuration(s); err == nil {
		return s, nil
	}

	// RFC3339 / RFC3339Nano
	if _, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return s, nil
	}

	// date format: YYYY-MM-DD
	if _, err := time.Parse("2006-01-02", s); err == nil {
		return s, nil
	}

	// unix timestamp
	if _, ok := parseUnixTimestamp(s); ok {
		return s, nil
	}

	return "", errors.New(
		"invalid since value (supported: duration, unix timestamp, date, RFC3339)",
	)
}

func parseUnixTimestamp(s string) (time.Time, bool) {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, false
	}

	switch len(s) {
	case 10: // seconds
		return time.Unix(n, 0), true
	case 13: // milliseconds
		return time.UnixMilli(n), true
	case 16: // microseconds
		return time.UnixMicro(n), true
	case 19: // nanoseconds
		return time.Unix(0, n), true
	default:
		return time.Time{}, false
	}
}

func validateKey(k string) (string, error) {
	if k == "" {
		return "", nil
	}
	if !keyPattern.MatchString(k) {
		return "", errors.New("invalid key")
	}
	return k, nil
}

func validateIdentifierList(s string, field string) (string, error) {
	if s == "" {
		return "", nil
	}
	if !identifierListPattern.MatchString(s) {
		return "", fmt.Errorf("invalid %s format", field)
	}
	return s, nil
}

func validateQuery(v string) string {
	v = strings.TrimSpace(v)

	// prevent accidental CLI flag injection
	v = strings.ReplaceAll(v, "\x00", "")

	// cap size
	if len(v) > 1000 {
		return v[:1000]
	}

	return v
}
