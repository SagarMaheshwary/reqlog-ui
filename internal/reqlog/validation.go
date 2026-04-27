package reqlog

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var keyPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
var servicePattern = regexp.MustCompile(`^[a-zA-Z0-9\-\*,]+$`)

func validateDir(dir string) (string, error) {
	clean := filepath.Clean(dir)

	absDir, err := filepath.Abs(clean)
	if err != nil {
		return "", errors.New("invalid directory path")
	}

	info, err := os.Stat(absDir)
	if err != nil || !info.IsDir() {
		return "", errors.New("directory does not exist")
	}

	return absDir, nil
}

func validateLimit(s string, defaultVal int, max int) int {
	limit, err := strconv.Atoi(s)
	if err != nil || limit < 0 || limit > max {
		return defaultVal
	}
	return limit
}

func validateSince(s string) (string, error) {
	if s == "" {
		return "", nil
	}
	_, err := time.ParseDuration(s)
	if err != nil {
		return "", errors.New("invalid since duration")
	}
	return s, nil
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

func validateService(s string) (string, error) {
	if s == "" {
		return "", nil
	}
	if !servicePattern.MatchString(s) {
		return "", errors.New("invalid service format")
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
