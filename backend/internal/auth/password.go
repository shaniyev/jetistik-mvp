package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

const bcryptCost = 12

// HashPassword creates a bcrypt hash of the given password.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword checks a password against a stored hash.
// It supports both bcrypt hashes and Django PBKDF2 hashes
// (format: pbkdf2_sha256$<iterations>$<salt>$<hash>).
// Returns (matches bool, needsRehash bool).
func VerifyPassword(password, storedHash string) (bool, bool) {
	// Try bcrypt first
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err == nil {
		return true, false
	}

	// Try Django PBKDF2
	if strings.HasPrefix(storedHash, "pbkdf2_sha256$") {
		if verifyDjangoPBKDF2(password, storedHash) {
			return true, true // matches, but needs rehash to bcrypt
		}
	}

	return false, false
}

// verifyDjangoPBKDF2 verifies a password against a Django-format PBKDF2 hash.
// Django format: pbkdf2_sha256$<iterations>$<salt>$<base64_hash>
func verifyDjangoPBKDF2(password, djangoHash string) bool {
	parts := strings.SplitN(djangoHash, "$", 4)
	if len(parts) != 4 {
		return false
	}

	algorithm := parts[0]
	if algorithm != "pbkdf2_sha256" {
		return false
	}

	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations <= 0 {
		return false
	}

	salt := parts[2]
	expectedHash := parts[3]

	derived := pbkdf2.Key([]byte(password), []byte(salt), iterations, sha256.Size, sha256.New)
	computedHash := base64.StdEncoding.EncodeToString(derived)

	return computedHash == expectedHash
}
