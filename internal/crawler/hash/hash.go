// Package hash provides the content-hashing function shared by url
// visit-state tracking and output record transformation.
package hash

import (
	"crypto/sha256"
	"encoding/hex"
)

// Content returns the hex-encoded SHA-256 hash of content.
func Content(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}
