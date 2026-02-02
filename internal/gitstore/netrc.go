package gitstore

import (
	"os"
	"path/filepath"
	"strings"
)

type netrcEntry struct {
	Login    string
	Password string
}

func netrcCredentials(host string) (string, string, bool, error) {
	path, ok := netrcPath()
	if !ok {
		return "", "", false, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", false, nil
		}
		return "", "", false, err
	}

	entries, fallback := parseNetrc(string(data))
	key := strings.ToLower(host)
	if entry, ok := entries[key]; ok {
		if entry.Login != "" || entry.Password != "" {
			return entry.Login, entry.Password, true, nil
		}
	}

	if fallback.Login != "" || fallback.Password != "" {
		return fallback.Login, fallback.Password, true, nil
	}

	return "", "", false, nil
}

func netrcPath() (string, bool) {
	if custom := os.Getenv("NETRC"); custom != "" {
		return custom, true
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".netrc"), true
	}
	return "", false
}

func parseNetrc(contents string) (map[string]netrcEntry, netrcEntry) {
	entries := map[string]netrcEntry{}
	defaultEntry := netrcEntry{}
	current := ""

	tokens := tokenizeNetrc(contents)
	for i := 0; i < len(tokens); i++ {
		token := strings.ToLower(tokens[i])
		switch token {
		case "machine":
			if i+1 < len(tokens) {
				current = strings.ToLower(tokens[i+1])
				i++
			}
		case "default":
			current = "default"
		case "login":
			if i+1 < len(tokens) {
				value := tokens[i+1]
				i++
				if current == "default" {
					defaultEntry.Login = value
				} else if current != "" {
					entry := entries[current]
					entry.Login = value
					entries[current] = entry
				}
			}
		case "password":
			if i+1 < len(tokens) {
				value := tokens[i+1]
				i++
				if current == "default" {
					defaultEntry.Password = value
				} else if current != "" {
					entry := entries[current]
					entry.Password = value
					entries[current] = entry
				}
			}
		}
	}

	return entries, defaultEntry
}

func tokenizeNetrc(contents string) []string {
	tokens := []string{}
	var current strings.Builder
	inQuote := false

	flush := func() {
		if current.Len() == 0 {
			return
		}
		tokens = append(tokens, current.String())
		current.Reset()
	}

	for i := 0; i < len(contents); i++ {
		char := contents[i]
		if char == '#' && !inQuote {
			for i < len(contents) && contents[i] != '\n' {
				i++
			}
			flush()
			continue
		}
		switch char {
		case ' ', '\t', '\n', '\r':
			if inQuote {
				current.WriteByte(char)
			} else {
				flush()
			}
		case '"':
			if inQuote {
				inQuote = false
				flush()
			} else {
				flush()
				inQuote = true
			}
		default:
			current.WriteByte(char)
		}
	}
	flush()

	return tokens
}
