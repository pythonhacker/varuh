package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"varuh"
)

// TestMapString tests the MapString function
func TestMapString(t *testing.T) {
	// Test case 1: Empty slice
	result := varuh.MapString([]string{}, func(s string) string {
		return strings.ToUpper(s)
	})
	if len(result) != 0 {
		t.Errorf("Expected empty slice, got %v", result)
	}

	// Test case 2: Single element
	result = varuh.MapString([]string{"hello"}, func(s string) string {
		return strings.ToUpper(s)
	})
	if len(result) != 1 || result[0] != "HELLO" {
		t.Errorf("Expected [HELLO], got %v", result)
	}

	// Test case 3: Multiple elements
	result = varuh.MapString([]string{"hello", "world", "test"}, func(s string) string {
		return strings.ToUpper(s)
	})
	expected := []string{"HELLO", "WORLD", "TEST"}
	if len(result) != 3 {
		t.Errorf("Expected length 3, got %d", len(result))
	}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("Expected %s at index %d, got %s", v, i, result[i])
		}
	}

	// Test case 4: Identity function
	result = varuh.MapString([]string{"a", "b", "c"}, func(s string) string {
		return s
	})
	expected = []string{"a", "b", "c"}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("Expected %s at index %d, got %s", v, i, result[i])
		}
	}
}

// TestHideSecret tests the HideSecret function
func TestHideSecret(t *testing.T) {
	// Test case 1: Empty string
	result := varuh.HideSecret("")
	if result != "" {
		t.Errorf("Expected empty string, got %s", result)
	}

	// Test case 2: Single character
	result = varuh.HideSecret("a")
	if result != "*" {
		t.Errorf("Expected '*', got %s", result)
	}

	// Test case 3: Multiple characters
	result = varuh.HideSecret("password")
	if len(result) != 8 {
		t.Errorf("Expected length 8, got %d", len(result))
	}
	for _, char := range result {
		if char != '*' {
			t.Errorf("Expected all '*' characters, got %c", char)
		}
	}

	// Test case 4: Long string
	longString := "this_is_a_very_long_password_string"
	result = varuh.HideSecret(longString)
	if len(result) != len(longString) {
		t.Errorf("Expected length %d, got %d", len(longString), len(result))
	}
	for _, char := range result {
		if char != '*' {
			t.Errorf("Expected all '*' characters, got %c", char)
		}
	}
}

// TestWriteSettings tests the WriteSettings function
func TestWriteSettings(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.json")

	// Test case 1: Valid settings
	settings := &varuh.Settings{
		ActiveDB:      "/tmp/test.db",
		Cipher:        "aes",
		AutoEncrypt:   true,
		KeepEncrypted: true,
		ShowPasswords: false,
		ConfigPath:    configFile,
		ListOrder:     "id,asc",
		Delim:         ">",
		Color:         "default",
		BgColor:       "bgblack",
	}

	err := varuh.WriteSettings(settings, configFile)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Errorf("Config file was not created")
	}

	// Test case 2: Invalid path (directory doesn't exist)
	invalidPath := "/nonexistent/path/config.json"
	err = varuh.WriteSettings(settings, invalidPath)
	if err == nil {
		t.Errorf("Expected error for invalid path, got nil")
	}
}

// TestUpdateSettings tests the UpdateSettings function
func TestUpdateSettings(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.json")

	// First create a config file
	settings := &varuh.Settings{
		ActiveDB:      "/tmp/test.db",
		Cipher:        "aes",
		AutoEncrypt:   true,
		KeepEncrypted: true,
		ShowPasswords: false,
		ConfigPath:    configFile,
		ListOrder:     "id,asc",
		Delim:         ">",
		Color:         "default",
		BgColor:       "bgblack",
	}

	err := varuh.WriteSettings(settings, configFile)
	if err != nil {
		t.Fatalf("Failed to create initial config file: %v", err)
	}

	// Test case 1: Update existing file
	settings.ShowPasswords = true
	settings.Color = "red"
	err = varuh.UpdateSettings(settings, configFile)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test case 2: Update non-existent file
	nonExistentFile := filepath.Join(tempDir, "nonexistent.json")
	err = varuh.UpdateSettings(settings, nonExistentFile)
	if err == nil {
		t.Errorf("Expected error for non-existent file, got nil")
	}
}

// TestGetColor tests the GetColor function
func TestGetColor(t *testing.T) {
	// Test case 1: Valid colors
	testCases := map[string]string{
		"black":   "\x1b[30m",
		"blue":    "\x1B[34m",
		"red":     "\x1B[31m",
		"green":   "\x1B[32m",
		"yellow":  "\x1B[33m",
		"magenta": "\x1B[35m",
		"cyan":    "\x1B[36m",
		"white":   "\x1B[37m",
		"bright":  "\x1b[1m",
		"dim":     "\x1b[2m",
		"reset":   "\x1B[0m",
		"default": "\x1B[0m",
	}

	for color, expected := range testCases {
		result := varuh.GetColor(color)
		if result != expected {
			t.Errorf("Expected %q for color %s, got %q", expected, color, result)
		}
	}

	// Test case 2: Invalid color (should return default)
	result := varuh.GetColor("invalid")
	expected := "\x1B[0m"
	if result != expected {
		t.Errorf("Expected default color for invalid input, got %q", result)
	}

	// Test case 3: Empty string
	result = varuh.GetColor("")
	if result != expected {
		t.Errorf("Expected default color for empty input, got %q", result)
	}
}

// TestPrettifyCardNumber tests the PrettifyCardNumber function
func TestPrettifyCardNumber(t *testing.T) {
	// Test case 1: 15-digit card (Amex format)
	result := varuh.PrettifyCardNumber("123456789012345")
	expected := "1234 567890 12345"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// Test case 2: 16-digit card (Standard format)
	result = varuh.PrettifyCardNumber("1234567890123456")
	expected = "1234 5678 9012 3456"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// Test case 3: Card with spaces (should remove spaces)
	result = varuh.PrettifyCardNumber("1234 5678 9012 3456")
	expected = "1234 5678 9012 3456"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// Test case 4: Invalid length (should return empty)
	result = varuh.PrettifyCardNumber("123")
	expected = ""
	if result != expected {
		t.Errorf("Expected empty string for invalid length, got %s", result)
	}

	// Test case 5: Empty string
	result = varuh.PrettifyCardNumber("")
	expected = ""
	if result != expected {
		t.Errorf("Expected empty string, got %s", result)
	}
}

// TestValidateCvv tests the ValidateCvv function
func TestValidateCvv(t *testing.T) {
	// Test case 1: Valid 3-digit CVV for non-Amex
	result := varuh.ValidateCvv("123", "Visa")
	if !result {
		t.Errorf("Expected true for valid 3-digit CVV, got false")
	}

	// Test case 2: Valid 4-digit CVV for Amex
	result = varuh.ValidateCvv("1234", "American Express")
	if !result {
		t.Errorf("Expected true for valid 4-digit CVV for Amex, got false")
	}

	// Test case 3: Invalid 3-digit CVV for Amex
	result = varuh.ValidateCvv("123", "American Express")
	if result {
		t.Errorf("Expected false for 3-digit CVV for Amex, got true")
	}

	// Test case 4: Invalid 4-digit CVV for non-Amex
	result = varuh.ValidateCvv("1234", "Visa")
	if result {
		t.Errorf("Expected false for 4-digit CVV for non-Amex, got true")
	}

	// Test case 5: Invalid CVV with letters
	result = varuh.ValidateCvv("abc", "Visa")
	if result {
		t.Errorf("Expected false for CVV with letters, got true")
	}

	// Test case 6: Empty CVV
	result = varuh.ValidateCvv("", "Visa")
	if result {
		t.Errorf("Expected false for empty CVV, got true")
	}
}

// TestValidateCardPin tests the ValidateCardPin function
func TestValidateCardPin(t *testing.T) {
	// Test case 1: Valid 4-digit PIN
	result := varuh.ValidateCardPin("1234")
	if !result {
		t.Errorf("Expected true for valid 4-digit PIN, got false")
	}

	// Test case 2: Valid 6-digit PIN
	result = varuh.ValidateCardPin("123456")
	if !result {
		t.Errorf("Expected true for valid 6-digit PIN, got false")
	}

	// Test case 3: Invalid 3-digit PIN
	result = varuh.ValidateCardPin("123")
	if result {
		t.Errorf("Expected false for 3-digit PIN, got true")
	}

	// Test case 4: Invalid PIN with letters
	result = varuh.ValidateCardPin("abcd")
	if result {
		t.Errorf("Expected false for PIN with letters, got true")
	}

	// Test case 5: Empty PIN
	result = varuh.ValidateCardPin("")
	if result {
		t.Errorf("Expected false for empty PIN, got true")
	}
}

// TestCheckValidExpiry tests the CheckValidExpiry function
func TestCheckValidExpiry(t *testing.T) {
	// Test case 1: Valid expiry date (current year)
	currentYear := time.Now().Year() % 100
	validExpiry := fmt.Sprintf("12/%d", currentYear)
	result := varuh.CheckValidExpiry(validExpiry)
	if !result {
		t.Errorf("Expected true for valid expiry %s, got false", validExpiry)
	}

	// Test case 2: Valid expiry date (future year)
	futureYear := currentYear + 1
	validExpiry = fmt.Sprintf("06/%d", futureYear)
	result = varuh.CheckValidExpiry(validExpiry)
	if !result {
		t.Errorf("Expected true for valid expiry %s, got false", validExpiry)
	}

	// Test case 3: Invalid month (0)
	result = varuh.CheckValidExpiry("0/25")
	if result {
		t.Errorf("Expected false for invalid month 0, got true")
	}

	// Test case 4: Invalid month (13)
	result = varuh.CheckValidExpiry("13/25")
	if result {
		t.Errorf("Expected false for invalid month 13, got true")
	}

	// Test case 5: Past year
	pastYear := currentYear - 1
	result = varuh.CheckValidExpiry(fmt.Sprintf("12/%d", pastYear))
	if result {
		t.Errorf("Expected false for past year, got true")
	}

	// Test case 6: Invalid format (single number)
	result = varuh.CheckValidExpiry("12")
	if result {
		t.Errorf("Expected false for invalid format, got true")
	}

	// Test case 7: Invalid format (no slash)
	result = varuh.CheckValidExpiry("1225")
	if result {
		t.Errorf("Expected false for invalid format, got true")
	}

	// Test case 8: Non-numeric month
	result = varuh.CheckValidExpiry("ab/25")
	if result {
		t.Errorf("Expected false for non-numeric month, got true")
	}

	// Test case 9: Non-numeric year
	result = varuh.CheckValidExpiry("12/ab")
	if result {
		t.Errorf("Expected false for non-numeric year, got true")
	}
}

// TestDetectCardType tests the DetectCardType function
func TestDetectCardType(t *testing.T) {
	// Test case 1: Valid Visa card
	cardType, err := varuh.DetectCardType("4111111111111111")
	if err != nil {
		t.Errorf("Expected no error for valid Visa card, got %v", err)
	}
	if cardType == "" {
		t.Errorf("Expected card type for Visa card, got empty string")
	}

	// Test case 2: Valid Mastercard
	cardType, err = varuh.DetectCardType("5555555555554444")
	if err != nil {
		t.Errorf("Expected no error for valid Mastercard, got %v", err)
	}
	if cardType == "" {
		t.Errorf("Expected card type for Mastercard, got empty string")
	}

	// Test case 3: Valid American Express
	cardType, err = varuh.DetectCardType("378282246310005")
	if err != nil {
		t.Errorf("Expected no error for valid Amex card, got %v", err)
	}
	if cardType == "" {
		t.Errorf("Expected card type for Amex card, got empty string")
	}

	// Test case 4: Empty card number
	cardType, err = varuh.DetectCardType("")
	if err == nil {
		t.Errorf("Expected error for empty card number, got nil")
	}

	// Test case 5: Invalid card number
	cardType, err = varuh.DetectCardType("1234567890123456")
	if err == nil {
		t.Errorf("Expected error for invalid card number, got nil")
	}
}

// TestRandomFileName tests the RandomFileName function
func TestRandomFileName(t *testing.T) {
	// Test case 1: Basic functionality
	folder := "/tmp"
	suffix := ".txt"
	result := varuh.RandomFileName(folder, suffix)

	// Check that result contains the folder path
	if !strings.HasPrefix(result, folder) {
		t.Errorf("Expected result to start with %s, got %s", folder, result)
	}

	// Check that result ends with suffix
	if !strings.HasSuffix(result, suffix) {
		t.Errorf("Expected result to end with %s, got %s", suffix, result)
	}

	// Check that the filename part is 32 characters (16 bytes * 2 for hex)
	baseName := filepath.Base(result)
	expectedLength := len(suffix) + 32 // 32 hex chars + suffix
	if len(baseName) != expectedLength {
		t.Errorf("Expected filename length %d, got %d", expectedLength, len(baseName))
	}

	// Test case 2: Empty folder
	result = varuh.RandomFileName("", suffix)
	if result == "" {
		t.Errorf("Expected non-empty result for empty folder")
	}

	// Test case 3: Empty suffix
	result = varuh.RandomFileName(folder, "")
	if !strings.HasPrefix(result, folder) {
		t.Errorf("Expected result to start with %s, got %s", folder, result)
	}
}

// TestSetShowPasswords tests the SetShowPasswords function
func TestSetShowPasswords(t *testing.T) {
	err := varuh.SetShowPasswords()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	// Note: We can't easily test the global state change without more complex setup
}

// TestSetCopyPasswordToClipboard tests the SetCopyPasswordToClipboard function
func TestSetCopyPasswordToClipboard(t *testing.T) {
	err := varuh.SetCopyPasswordToClipboard()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	// Note: We can't easily test the global state change without more complex setup
}

// TestSetAssumeYes tests the SetAssumeYes function
func TestSetAssumeYes(t *testing.T) {
	err := varuh.SetAssumeYes()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	// Note: We can't easily test the global state change without more complex setup
}

// TestSetType tests the SetType function
func TestSetType(t *testing.T) {
	testType := "card"
	varuh.SetType(testType)
	// Note: We can't easily test the global state change without more complex setup
}

// TestCopyPasswordToClipboard tests the CopyPasswordToClipboard function
func TestCopyPasswordToClipboard(t *testing.T) {
	testPassword := "test_password_123"
	varuh.CopyPasswordToClipboard(testPassword)
	// Note: We can't easily test clipboard functionality without more complex setup
	// This test mainly ensures the function doesn't panic
}

// TestRewriteBaseFile tests the RewriteBaseFile function
func TestRewriteBaseFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt.enc")
	contents := []byte("test content")
	mode := os.FileMode(0644)

	// Test case 1: Valid file creation
	err, resultPath := varuh.RewriteBaseFile(testFile, contents, mode)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedPath := filepath.Join(tempDir, "test.txt")
	if resultPath != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, resultPath)
	}

	// Verify file was created
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("File was not created")
	}

	// Verify file contents
	readContents, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Errorf("Failed to read created file: %v", err)
	}
	if string(readContents) != string(contents) {
		t.Errorf("Expected contents %s, got %s", string(contents), string(readContents))
	}
}

// TestRewriteFile tests the RewriteFile function
func TestRewriteFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	contents := []byte("test content")
	mode := os.FileMode(0644)

	// Test case 1: Valid file creation
	err, resultPath := varuh.RewriteFile(testFile, contents, mode)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if resultPath != testFile {
		t.Errorf("Expected path %s, got %s", testFile, resultPath)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Errorf("File was not created")
	}

	// Verify file contents
	readContents, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("Failed to read created file: %v", err)
	}
	if string(readContents) != string(contents) {
		t.Errorf("Expected contents %s, got %s", string(contents), string(readContents))
	}
}

// TestPrintDelim tests the PrintDelim function
func TestPrintDelim(t *testing.T) {
	// This function prints to stdout, so we mainly test that it doesn't panic
	// We can't easily capture stdout in a unit test without complex setup

	// Test case 1: Normal delimiter
	varuh.PrintDelim(">", "default")

	// Test case 2: Underscore color (should change delimiter to space)
	varuh.PrintDelim(">", "underscore")

	// Test case 3: Multi-character delimiter (should take first character)
	varuh.PrintDelim(">>>", "default")

	// Test case 4: Empty delimiter
	varuh.PrintDelim("", "default")

	// If we get here without panicking, the test passes
}

// TestGetOrCreateLocalConfig tests the GetOrCreateLocalConfig function
func TestGetOrCreateLocalConfig(t *testing.T) {
	// Test case 1: Create config for test app
	appName := "test_app_" + fmt.Sprintf("%d", time.Now().Unix())
	err, settings := varuh.GetOrCreateLocalConfig(appName)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if settings == nil {
		t.Errorf("Expected settings, got nil")
	}

	// Verify default settings
	if settings.Cipher != "aes" {
		t.Errorf("Expected cipher 'aes', got %s", settings.Cipher)
	}
	if !settings.AutoEncrypt {
		t.Errorf("Expected AutoEncrypt to be true, got false")
	}
	if !settings.KeepEncrypted {
		t.Errorf("Expected KeepEncrypted to be true, got false")
	}
	if settings.ShowPasswords {
		t.Errorf("Expected ShowPasswords to be false, got true")
	}
	if settings.ListOrder != "id,asc" {
		t.Errorf("Expected ListOrder 'id,asc', got %s", settings.ListOrder)
	}
	if settings.Delim != ">" {
		t.Errorf("Expected Delim '>', got %s", settings.Delim)
	}
	if settings.Color != "default" {
		t.Errorf("Expected Color 'default', got %s", settings.Color)
	}
	if settings.BgColor != "bgblack" {
		t.Errorf("Expected BgColor 'bgblack', got %s", settings.BgColor)
	}

	// Test case 2: Get existing config
	err, settings2 := varuh.GetOrCreateLocalConfig(appName)
	if err != nil {
		t.Errorf("Expected no error for existing config, got %v", err)
	}
	if settings2 == nil {
		t.Errorf("Expected settings for existing config, got nil")
	}
}

// TestHasActiveDatabase tests the HasActiveDatabase function
func TestHasActiveDatabase(t *testing.T) {
	// This function depends on the global config and file system state
	// We can't easily test it without complex setup, but we can ensure it doesn't panic
	result := varuh.HasActiveDatabase()
	// result should be a boolean
	_ = result
}

// TestGetActiveDatabase tests the GetActiveDatabase function
func TestGetActiveDatabase(t *testing.T) {
	// This function depends on the global config and file system state
	// We can't easily test it without complex setup, but we can ensure it doesn't panic
	err, dbPath := varuh.GetActiveDatabase()
	// The function can return nil error and empty path when no active database is configured
	// This is a valid state, so we just ensure it doesn't panic
	_ = err
	_ = dbPath
}

// TestUpdateActiveDbPath tests the UpdateActiveDbPath function
func TestUpdateActiveDbPath(t *testing.T) {
	// This function depends on the global config and file system state
	// We can't easily test it without complex setup, but we can ensure it doesn't panic
	testPath := "/tmp/test.db"
	err := varuh.UpdateActiveDbPath(testPath)
	// We expect either an error or success
	_ = err
}
