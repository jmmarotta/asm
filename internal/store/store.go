package store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
)

func RepoKey(origin string, ref string) string {
	payload := fmt.Sprintf("%s|%s", origin, ref)
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

func RepoPath(storeDir string, origin string, ref string) string {
	return filepath.Join(storeDir, RepoKey(origin, ref))
}
