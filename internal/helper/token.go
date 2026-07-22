package helper

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/ssh"
)

// GenerateToken returns a high-entropy random token (hex) and its prefix
// (first 8 chars) used for display and lookup.
func GenerateToken() (plain string, prefix string, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", "", err
	}
	plain = hex.EncodeToString(b)
	prefix = plain[:8]
	hash = HashToken(plain)
	return plain, prefix, hash, nil
}

// HashToken hashes a plaintext token with SHA-256 for storage.
func HashToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

// FingerprintSSHKey computes an OpenSSH-style MD5 fingerprint of a public key,
// e.g. "SHA256:abcd...". Accepts the raw authorized_keys line.
func FingerprintSSHKey(publicKey string) (string, error) {
	pub, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
	if err != nil {
		return "", fmt.Errorf("invalid ssh public key: %w", err)
	}
	// OpenSSH SHA256 fingerprint (no trailing algorithm label).
	fp := ssh.FingerprintSHA256(pub)
	// fp is "SHA256:xxxx"; return as-is for clarity.
	return fp, nil
}
