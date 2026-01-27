package debug

import (
	"io"
	"log"
	"net/url"
	"strings"
	"sync"
)

var (
	mu      sync.RWMutex
	enabled bool
	logger  *log.Logger
)

// Configure enables or disables debug logging.
func Configure(on bool, out io.Writer) {
	mu.Lock()
	defer mu.Unlock()

	enabled = on
	if !on || out == nil {
		logger = nil
		return
	}

	logger = log.New(out, "debug: ", log.LstdFlags|log.Lmicroseconds)
}

// Enabled reports whether debug logging is enabled.
func Enabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return enabled
}

// Logf writes a formatted debug message when enabled.
func Logf(format string, args ...any) {
	mu.RLock()
	current := logger
	on := enabled
	mu.RUnlock()
	if !on || current == nil {
		return
	}

	current.Printf(format, args...)
}

// SanitizeOrigin removes credential-like data from origins for logs.
func SanitizeOrigin(origin string) string {
	if origin == "" {
		return origin
	}
	if strings.Contains(origin, "://") {
		parsed, err := url.Parse(origin)
		if err != nil {
			return origin
		}
		parsed.User = nil
		return parsed.String()
	}
	if at := strings.Index(origin, "@"); at != -1 {
		if strings.Contains(origin[at+1:], ":") {
			return origin[at+1:]
		}
	}
	return origin
}
