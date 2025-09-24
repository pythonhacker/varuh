package tests

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"varuh"
)

func TestGenerateRandomBytes(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		wantErr  bool
		checkLen bool
	}{
		{"valid small size", 16, false, true},
		{"valid medium size", 32, false, true},
		{"valid large size", 128, false, true},
		{"zero size", 0, false, true},
		{"negative size", -1, false, true}, // Go's make will panic on negative size
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.size < 0 {
				defer func() {
					if r := recover(); r == nil {
						t.Error("Expected panic for negative size")
					}
				}()
			}

			err, data := varuh.GenerateRandomBytes(tt.size)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateRandomBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkLen && len(data) != tt.size {
				t.Errorf("GenerateRandomBytes() returned data of length %d, want %d", len(data), tt.size)
			}

			// Check that data is actually random (not all zeros)
			if tt.size > 0 && !tt.wantErr {
				allZeros := true
				for _, b := range data {
					if b != 0 {
						allZeros = false
						break
					}
				}
				// Very unlikely to get all zeros with crypto/rand
				if allZeros {
					t.Error("GenerateRandomBytes() returned all zeros, likely not random")
				}
			}
		})
	}
}

func TestGenerateKeyArgon2(t *testing.T) {
	tests := []struct {
		name       string
		passPhrase string
		salt       *[]byte
		wantErr    bool
		checkSalt  bool
	}{
		{"valid passphrase no salt", "test password", nil, false, true},
		{"valid passphrase with salt", "test password", func() *[]byte { s := make([]byte, varuh.SALT_SIZE); return &s }(), false, false},
		{"empty passphrase", "", nil, false, true},
		{"invalid salt size", "test", func() *[]byte { s := make([]byte, 10); return &s }(), true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, key, salt := varuh.GenerateKeyArgon2(tt.passPhrase, tt.salt)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateKeyArgon2() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(key) != varuh.KEY_SIZE {
					t.Errorf("GenerateKeyArgon2() returned key of length %d, want %d", len(key), varuh.KEY_SIZE)
				}

				if tt.checkSalt && len(salt) != varuh.SALT_SIZE {
					t.Errorf("GenerateKeyArgon2() returned salt of length %d, want %d", len(salt), varuh.SALT_SIZE)
				}

				// Test deterministic key generation with same salt
				if tt.salt != nil {
					err2, key2, _ := varuh.GenerateKeyArgon2(tt.passPhrase, tt.salt)
					if err2 != nil {
						t.Errorf("Second GenerateKeyArgon2() call failed: %v", err2)
					} else if !bytes.Equal(key, key2) {
						t.Error("GenerateKeyArgon2() should produce same key with same salt")
					}
				}
			}
		})
	}
}

func TestGenerateKey(t *testing.T) {
	tests := []struct {
		name       string
		passPhrase string
		salt       *[]byte
		wantErr    bool
		checkSalt  bool
	}{
		{"valid passphrase no salt", "test password", nil, false, true},
		{"valid passphrase with salt", "test password", func() *[]byte { s := make([]byte, varuh.SALT_SIZE); return &s }(), false, false},
		{"empty passphrase", "", nil, false, true},
		{"invalid salt size", "test", func() *[]byte { s := make([]byte, 10); return &s }(), true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, key, salt := varuh.GenerateKey(tt.passPhrase, tt.salt)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(key) != varuh.KEY_SIZE {
					t.Errorf("GenerateKey() returned key of length %d, want %d", len(key), varuh.KEY_SIZE)
				}

				if tt.checkSalt && len(salt) != varuh.SALT_SIZE {
					t.Errorf("GenerateKey() returned salt of length %d, want %d", len(salt), varuh.SALT_SIZE)
				}

				// Test deterministic key generation with same salt
				if tt.salt != nil {
					err2, key2, _ := varuh.GenerateKey(tt.passPhrase, tt.salt)
					if err2 != nil {
						t.Errorf("Second GenerateKey() call failed: %v", err2)
					} else if !bytes.Equal(key, key2) {
						t.Error("GenerateKey() should produce same key with same salt")
					}
				}
			}
		})
	}
}

func TestIsFileEncrypted(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	encryptedFile := filepath.Join(tempDir, "encrypted.db")
	unencryptedFile := filepath.Join(tempDir, "unencrypted.db")
	emptyFile := filepath.Join(tempDir, "empty.db")
	nonExistentFile := filepath.Join(tempDir, "nonexistent.db")

	// Create encrypted file with magic header
	magicBytes := []byte(fmt.Sprintf("%x", varuh.MAGIC_HEADER))
	err := os.WriteFile(encryptedFile, append(magicBytes, []byte("some encrypted data")...), 0644)
	if err != nil {
		t.Fatalf("Failed to create test encrypted file: %v", err)
	}

	// Create unencrypted file
	err = os.WriteFile(unencryptedFile, []byte("regular data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test unencrypted file: %v", err)
	}

	// Create empty file
	err = os.WriteFile(emptyFile, []byte{}, 0644)
	if err != nil {
		t.Fatalf("Failed to create test empty file: %v", err)
	}

	tests := []struct {
		name      string
		filePath  string
		wantErr   bool
		encrypted bool
	}{
		{"encrypted file", encryptedFile, false, true},
		{"unencrypted file", unencryptedFile, true, false},
		{"empty file", emptyFile, true, false},
		{"non-existent file", nonExistentFile, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, encrypted := varuh.IsFileEncrypted(tt.filePath)

			if (err != nil) != tt.wantErr {
				t.Errorf("IsFileEncrypted() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if encrypted != tt.encrypted {
				t.Errorf("IsFileEncrypted() encrypted = %v, want %v", encrypted, tt.encrypted)
			}
		})
	}
}

func TestEncryptFileAES(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.db")
	testContent := []byte("This is test database content for AES encryption")

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		dbPath   string
		password string
		wantErr  bool
	}{
		{"valid encryption", testFile, "testpassword", false},
		{"empty password", testFile, "", false}, // Empty password should still work
		{"non-existent file", filepath.Join(tempDir, "nonexistent.db"), "password", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := varuh.EncryptFileAES(tt.dbPath, tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("EncryptFileAES() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the file is now encrypted
				err, encrypted := varuh.IsFileEncrypted(tt.dbPath)
				if err != nil {
					t.Errorf("Failed to check if file is encrypted: %v", err)
				} else if !encrypted {
					t.Error("File should be encrypted after EncryptFileAES()")
				}
			}
		})
	}
}

func TestDecryptFileAES(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.db")
	testContent := []byte("This is test database content for AES decryption")
	password := "testpassword"

	// Create and encrypt a test file
	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = varuh.EncryptFileAES(testFile, password)
	if err != nil {
		t.Fatalf("Failed to encrypt test file: %v", err)
	}

	tests := []struct {
		name     string
		filePath string
		password string
		wantErr  bool
	}{
		{"valid decryption", testFile, password, false},
		{"wrong password", testFile, "wrongpassword", true},
		{"non-existent file", filepath.Join(tempDir, "nonexistent.db"), password, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Re-encrypt file for each test case
			if tt.name != "non-existent file" {
				os.WriteFile(testFile, testContent, 0644)
				varuh.EncryptFileAES(testFile, password)
			}

			err := varuh.DecryptFileAES(tt.filePath, tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptFileAES() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the file is decrypted and content is correct
				decryptedContent, err := os.ReadFile(tt.filePath)
				if err != nil {
					t.Errorf("Failed to read decrypted file: %v", err)
				} else if !bytes.Equal(decryptedContent, testContent) {
					t.Error("Decrypted content doesn't match original content")
				}
			}
		})
	}
}

func TestEncryptFileXChachaPoly(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.db")
	testContent := []byte("This is test database content for XChaCha encryption")

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		dbPath   string
		password string
		wantErr  bool
	}{
		{"valid encryption", testFile, "testpassword", false},
		{"empty password", testFile, "", false},
		{"non-existent file", filepath.Join(tempDir, "nonexistent.db"), "password", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := varuh.EncryptFileXChachaPoly(tt.dbPath, tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("EncryptFileXChachaPoly() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the file is now encrypted
				err, encrypted := varuh.IsFileEncrypted(tt.dbPath)
				if err != nil {
					t.Errorf("Failed to check if file is encrypted: %v", err)
				} else if !encrypted {
					t.Error("File should be encrypted after EncryptFileXChachaPoly()")
				}
			}
		})
	}
}

func TestDecryptFileXChachaPoly(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.db")
	testContent := []byte("This is test database content for XChaCha decryption")
	password := "testpassword"

	// Create and encrypt a test file
	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = varuh.EncryptFileXChachaPoly(testFile, password)
	if err != nil {
		t.Fatalf("Failed to encrypt test file: %v", err)
	}

	tests := []struct {
		name     string
		filePath string
		password string
		wantErr  bool
	}{
		{"valid decryption", testFile, password, false},
		{"wrong password", testFile, "wrongpassword", true},
		{"non-existent file", filepath.Join(tempDir, "nonexistent.db"), password, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Re-encrypt file for each test case
			if tt.name != "non-existent file" {
				os.WriteFile(testFile, testContent, 0644)
				varuh.EncryptFileXChachaPoly(testFile, password)
			}

			err := varuh.DecryptFileXChachaPoly(tt.filePath, tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptFileXChachaPoly() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the file is decrypted and content is correct
				decryptedContent, err := os.ReadFile(tt.filePath)
				if err != nil {
					t.Errorf("Failed to read decrypted file: %v", err)
				} else if !bytes.Equal(decryptedContent, testContent) {
					t.Error("Decrypted content doesn't match original content")
				}
			}
		})
	}
}

func TestGeneratePassword(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{"zero length", 0, false},
		{"small length", 8, false},
		{"medium length", 16, false},
		{"large length", 64, false},
		{"negative length", -1, false}, // This should handle gracefully or error
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.length < 0 {
				// Expected to either handle gracefully or panic
				defer func() {
					recover() // Catch any panic
				}()
			}

			err, password := varuh.GeneratePassword(tt.length)

			if (err != nil) != tt.wantErr {
				t.Errorf("GeneratePassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.length >= 0 {
				if len(password) != tt.length {
					t.Errorf("GeneratePassword() returned password of length %d, want %d", len(password), tt.length)
				}

				// Check that password contains only valid characters
				const validChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789=+_()$#@!~:/%"
				for _, char := range password {
					found := false
					for _, validChar := range validChars {
						if char == validChar {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("GeneratePassword() returned invalid character: %c", char)
					}
				}
			}
		})
	}
}

func TestGenerateStrongPassword(t *testing.T) {
	// Run multiple times since it's random
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i+1), func(t *testing.T) {
			err, password := varuh.GenerateStrongPassword()

			if err != nil {
				t.Errorf("GenerateStrongPassword() error = %v", err)
				return
			}

			// Check minimum length
			if len(password) < 12 {
				t.Errorf("GenerateStrongPassword() returned password of length %d, minimum expected 12", len(password))
			}

			// Check maximum expected length (should be 16 or less based on implementation)
			if len(password) > 16 {
				t.Errorf("GenerateStrongPassword() returned password of length %d, maximum expected 16", len(password))
			}

			// Check that it contains various character types
			hasLower := false
			hasUpper := false
			hasDigit := false
			hasPunct := false

			const lowerChars = "abcdefghijklmnopqrstuvwxyz"
			const upperChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
			const digitChars = "0123456789"
			const punctChars = "=+_()$#@!~:/%"

			for _, char := range password {
				switch {
				case containsChar(lowerChars, char):
					hasLower = true
				case containsChar(upperChars, char):
					hasUpper = true
				case containsChar(digitChars, char):
					hasDigit = true
				case containsChar(punctChars, char):
					hasPunct = true
				}
			}

			if !hasLower {
				t.Error("GenerateStrongPassword() should contain lowercase characters")
			}
			if !hasUpper {
				t.Error("GenerateStrongPassword() should contain uppercase characters")
			}
			if !hasDigit {
				t.Error("GenerateStrongPassword() should contain digit characters")
			}
			if !hasPunct {
				t.Error("GenerateStrongPassword() should contain punctuation characters")
			}
		})
	}
}

// Helper function to check if a string contains a character
func containsChar(s string, char rune) bool {
	for _, c := range s {
		if c == char {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkGenerateRandomBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		varuh.GenerateRandomBytes(32)
	}
}

func BenchmarkGenerateKeyArgon2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		varuh.GenerateKeyArgon2("test password", nil)
	}
}

func BenchmarkGenerateKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		varuh.GenerateKey("test password", nil)
	}
}
