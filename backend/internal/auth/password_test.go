package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"

	"golang.org/x/crypto/pbkdf2"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("testpassword123")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}
	if hash == "testpassword123" {
		t.Fatal("HashPassword returned plaintext")
	}
}

func TestVerifyPassword_Bcrypt(t *testing.T) {
	hash, err := HashPassword("mypassword")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	matches, needsRehash := VerifyPassword("mypassword", hash)
	if !matches {
		t.Error("expected password to match")
	}
	if needsRehash {
		t.Error("bcrypt hash should not need rehash")
	}

	matches, _ = VerifyPassword("wrongpassword", hash)
	if matches {
		t.Error("expected wrong password to not match")
	}
}

func TestVerifyPassword_DjangoPBKDF2(t *testing.T) {
	password := "testpass123"
	djangoHash := createTestDjangoHash(password, "testsalt123", 260000)

	matches, needsRehash := VerifyPassword(password, djangoHash)
	if !matches {
		t.Error("expected Django PBKDF2 password to match")
	}
	if !needsRehash {
		t.Error("Django PBKDF2 hash should need rehash to bcrypt")
	}

	matches, _ = VerifyPassword("wrongpassword", djangoHash)
	if matches {
		t.Error("expected wrong password to not match Django hash")
	}
}

func TestVerifyPassword_InvalidHash(t *testing.T) {
	matches, needsRehash := VerifyPassword("password", "invalid_hash")
	if matches {
		t.Error("expected invalid hash to not match")
	}
	if needsRehash {
		t.Error("expected invalid hash to not need rehash")
	}
}

func createTestDjangoHash(password, salt string, iterations int) string {
	derived := pbkdf2.Key([]byte(password), []byte(salt), iterations, sha256.Size, sha256.New)
	hash := base64.StdEncoding.EncodeToString(derived)
	return fmt.Sprintf("pbkdf2_sha256$%d$%s$%s", iterations, salt, hash)
}
