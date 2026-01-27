package store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
)

func RepoKey(origin string) string {
	payload := fmt.Sprintf("%s", origin)
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

func RepoPath(storeDir string, origin string) string {
	return filepath.Join(storeDir, RepoKey(origin))
}
