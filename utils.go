// Utility functions
package main

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/kirsle/configdir"
	"github.com/polyglothacker/creditcard"
	"golang.org/x/crypto/ssh/terminal"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const DELIMSIZE int = 69

// Over-ride settings via cmd line
type SettingsOverride struct {
	ShowPasswords bool
	CopyPassword  bool
	AssumeYes     bool
	Type          string // Type of entity to add
}

// Settings structure for local config
type Settings struct {
	ActiveDB      string `json:"active_db"`
	Cipher        string `json:"cipher"`
	AutoEncrypt   bool   `json:"auto_encrypt"`
	KeepEncrypted bool   `json:"encrypt_on"`
	ShowPasswords bool   `json:"visible_passwords"`
	ConfigPath    string `json:"path"`
	// Key to order listings when using -a option
	// Valid values are
	// 1. timestamp,{desc,asc}
	// 2. title,{desc,asc}
	// 3. username, {desc,asc}
	// 4. id, {desc,asc{
	ListOrder string `json:"list_order"`
	Delim     string `json:"delimiter"`
	Color     string `json:"color"`   // fg color to print
	BgColor   string `json:"bgcolor"` // bg color to print
}

// Global settings override
var settingsRider SettingsOverride

// Write settings to disk
func writeSettings(settings *Settings, configFile string) error {

	fh, err := os.Create(configFile)
	if err != nil {
		fmt.Printf("Error generating configuration file %s - \"%s\"\n", configFile, err.Error())
		return err
	}

	defer fh.Close()

	encoder := json.NewEncoder(fh)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(&settings)

	return err
}

// Write updated settings to disk
func updateSettings(settings *Settings, configFile string) error {

	fh, err := os.OpenFile(configFile, os.O_RDWR, 0644)
	if err != nil {
		fmt.Printf("Error opening config file %s - \"%s\"\n", configFile, err.Error())
		return err
	}

	defer fh.Close()

	encoder := json.NewEncoder(fh)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(&settings)

	if err != nil {
		fmt.Printf("Error updating config %s - \"%s\"\n", configFile, err.Error())
		return err
	}

	return err
}

// Make the per-user configuration folder and return local settings
func getOrCreateLocalConfig(app string) (error, *Settings) {

	var settings Settings
	var configPath string
	var configFile string
	var err error
	var fh *os.File

	configPath = configdir.LocalConfig(app)
	err = configdir.MakePath(configPath) // Ensure it exists.
	if err != nil {
		return err, nil
	}

	configFile = filepath.Join(configPath, "config.json")
	//  fmt.Printf("Config file, path => %s %s\n", configFile, configPath)

	if _, err = os.Stat(configFile); err == nil {
		fh, err = os.Open(configFile)
		if err != nil {
			return err, nil
		}

		defer fh.Close()

		decoder := json.NewDecoder(fh)
		err = decoder.Decode(&settings)
		if err != nil {
			return err, nil
		}

	} else {
		//      fmt.Printf("Creating default configuration ...")
		settings = Settings{"", "aes", true, true, false, configFile, "id,asc", ">", "default", "bgblack"}

		if err = writeSettings(&settings, configFile); err == nil {
			// fmt.Println(" ...done")
		} else {
			return err, nil
		}
	}

	return nil, &settings
}

// Return if there is an active, decrypted database
func hasActiveDatabase() bool {

	err, settings := getOrCreateLocalConfig(APP)
	if err == nil && settings.ActiveDB != "" {
		if _, err := os.Stat(settings.ActiveDB); err == nil {
			if _, flag := isFileEncrypted(settings.ActiveDB); !flag {
				return true
			}
			return false
		}
	}

	if err != nil {
		fmt.Printf("Error parsing local config - \"%s\"\n", err.Error())
	}

	return false
}

// Get the current active database
func getActiveDatabase() (error, string) {

	err, settings := getOrCreateLocalConfig(APP)
	if err == nil && settings.ActiveDB != "" {
		if _, err := os.Stat(settings.ActiveDB); err == nil {
			return nil, settings.ActiveDB
		}
	}

	if err != nil {
		fmt.Printf("Error parsing local config - \"%s\"\n", err.Error())
	}

	return err, ""
}

// Update the active db path
func updateActiveDbPath(dbPath string) error {

	_, settings := getOrCreateLocalConfig(APP)

	if settings != nil {
		settings.ActiveDB = dbPath
	}

	return updateSettings(settings, settings.ConfigPath)

}

// Read the password from console without echoing
func readPassword() (error, string) {

	var passwd []byte
	var err error

	passwd, err = terminal.ReadPassword(int(os.Stdin.Fd()))
	return err, string(passwd)
}

// Rewrite the contents of the base file (path minus extension) with the new contents
func rewriteBaseFile(path string, contents []byte, mode fs.FileMode) (error, string) {

	var err error
	var origFile string

	origFile = strings.TrimSuffix(path, filepath.Ext(path))
	// Overwrite it
	err = os.WriteFile(origFile, contents, 0644)

	if err == nil {
		// Chmod it
		os.Chmod(origFile, mode)
	}

	return err, origFile
}

// Rewrite the contents of the file with the new contents
func rewriteFile(path string, contents []byte, mode fs.FileMode) (error, string) {

	var err error

	// Overwrite it
	err = os.WriteFile(path, contents, 0644)

	if err == nil {
		// Chmod it
		os.Chmod(path, mode)
	}

	return err, path
}

// Get color codes for console colors
func getColor(code string) string {

	colors := map[string]string{
		"black":   "\x1b[30m",
		"blue":    "\x1B[34m",
		"red":     "\x1B[31m",
		"green":   "\x1B[32m",
		"yellow":  "\x1B[33m",
		"magenta": "\x1B[35m",
		"cyan":    "\x1B[36m",
		"white":   "\x1B[37m",

		// From https://gist.github.com/abritinthebay/d80eb99b2726c83feb0d97eab95206c4
		// esoteric options
		"bright":     "\x1b[1m",
		"dim":        "\x1b[2m",
		"underscore": "\x1b[4m",
		"blink":      "\x1b[5m",
		"reverse":    "\x1b[7m",
		"hidden":     "\x1b[8m",

		// background color options
		"bgblack":   "\x1b[40m",
		"bgred":     "\x1b[41m",
		"bggreen":   "\x1b[42m",
		"bgyellow":  "\x1b[43m",
		"bgblue":    "\x1b[44m",
		"bgmagenta": "\x1b[45m",
		"bgcyan":    "\x1b[46m",
		"bgwhite":   "\x1b[47m",

		// reset color code
		"reset":   "\x1B[0m",
		"default": "\x1B[0m",
	}

	if color, ok := colors[code]; ok {
		return color
	} else {
		return colors["default"]
	}

}

// Print the delimiter line for listings
func printDelim(delimChar string, color string) {

	var delims []string

	if color == "underscore" {
		// Override delimieter to space
		delimChar = " "
	}

	if len(delimChar) > 1 {
		// slice it - take only the first
		delimChar = string(delimChar[0])
	}
	for i := 0; i < DELIMSIZE; i++ {
		delims = append(delims, delimChar)
	}

	fmt.Println(strings.Join(delims, ""))
}

// Prettify credit/debit card numbers
func prettifyCardNumber(cardNumber string) string {

	// Amex cards are 15 digits - group as 4, 6, 5
	// Any 16 digits - group as 4/4/4/4
	var numbers []string

	if len(cardNumber) == 15 {
		numbers = append(numbers, cardNumber[0:4])
		numbers = append(numbers, cardNumber[4:10])
		numbers = append(numbers, cardNumber[10:15])
	} else if len(cardNumber) == 16 {
		numbers = append(numbers, cardNumber[0:4])
		numbers = append(numbers, cardNumber[4:8])
		numbers = append(numbers, cardNumber[8:12])
		numbers = append(numbers, cardNumber[12:16])
	}

	return strings.Join(numbers, " ")
}

// Print a card entry to the console
func printCardEntry(entry *Entry, settings *Settings, delim bool) error {

	var customEntries []ExtendedEntry

	fmt.Printf("%s", getColor(strings.ToLower(settings.Color)))
	if strings.HasPrefix(settings.BgColor, "bg") {
		fmt.Printf("%s", getColor(strings.ToLower(settings.BgColor)))
	}

	if delim {
		printDelim(settings.Delim, settings.Color)
	}

	fmt.Printf("[Type: card]\n")
	fmt.Printf("ID: %d\n", entry.ID)
	fmt.Printf("Card Name: %s\n", entry.Title)
	fmt.Printf("Card Holder: %s\n", entry.User)
	fmt.Printf("Card Number: %s\n", prettifyCardNumber(entry.Url))
	fmt.Printf("Card Type: %s\n", entry.Class)

	if entry.Issuer != "" {
		fmt.Printf("Issuing Bank: %s\n", entry.Issuer)
	}

	fmt.Println()

	fmt.Printf("Expiry Date: %s\n", entry.ExpiryDate)

	if settings.ShowPasswords || settingsRider.ShowPasswords {
		fmt.Printf("Card CVV: %s\n", entry.Password)
		fmt.Printf("Card PIN: %s\n", entry.Pin)
	} else {
		var asterisks1 []string
		var asterisks2 []string
		var i int

		for i = 0; i < len(entry.Password); i++ {
			asterisks1 = append(asterisks1, "*")
		}
		fmt.Printf("Card CVV: %s\n", strings.Join(asterisks1, ""))

		for i = 0; i < len(entry.Pin); i++ {
			asterisks2 = append(asterisks2, "*")
		}
		fmt.Printf("Card PIN: %s\n", strings.Join(asterisks2, ""))
	}

	if len(entry.Tags) > 0 {
		fmt.Printf("\nTags: %s\n", entry.Tags)
	}
	if len(entry.Notes) > 0 {
		fmt.Printf("Notes: %s\n", entry.Notes)
	}
	// Query extended entries
	customEntries = getExtendedEntries(entry)

	if len(customEntries) > 0 {
		for _, customEntry := range customEntries {
			fmt.Printf("%s: %s\n", customEntry.FieldName, customEntry.FieldValue)
		}
	}

	fmt.Printf("Modified: %s\n", entry.Timestamp.Format("2006-01-02 15:04:05"))

	printDelim(settings.Delim, settings.Color)

	// Reset
	fmt.Printf("%s", getColor("default"))

	return nil

}

// Print an entry to the console
func printEntry(entry *Entry, delim bool) error {

	var err error
	var settings *Settings
	var customEntries []ExtendedEntry

	err, settings = getOrCreateLocalConfig(APP)

	if err != nil {
		fmt.Printf("Error parsing config - \"%s\"\n", err.Error())
		return err
	}

	if entry.Type == "card" {
		return printCardEntry(entry, settings, delim)
	}

	fmt.Printf("%s", getColor(strings.ToLower(settings.Color)))
	if strings.HasPrefix(settings.BgColor, "bg") {
		fmt.Printf("%s", getColor(strings.ToLower(settings.BgColor)))
	}

	if delim {
		printDelim(settings.Delim, settings.Color)
	}

	fmt.Printf("[Type: password]\n")
	fmt.Printf("ID: %d\n", entry.ID)
	fmt.Printf("Title: %s\n", entry.Title)
	fmt.Printf("User: %s\n", entry.User)
	fmt.Printf("URL: %s\n", entry.Url)

	if settings.ShowPasswords || settingsRider.ShowPasswords {
		fmt.Printf("Password: %s\n", entry.Password)
	} else {
		var asterisks []string

		for i := 0; i < len(entry.Password); i++ {
			asterisks = append(asterisks, "*")
		}
		fmt.Printf("Password: %s\n", strings.Join(asterisks, ""))
	}

	if len(entry.Tags) > 0 {
		fmt.Printf("Tags: %s\n", entry.Tags)
	}
	if len(entry.Notes) > 0 {
		fmt.Printf("Notes: %s\n", entry.Notes)
	}
	// Query extended entries
	customEntries = getExtendedEntries(entry)

	if len(customEntries) > 0 {
		for _, customEntry := range customEntries {
			fmt.Printf("%s: %s\n", customEntry.FieldName, customEntry.FieldValue)
		}
	}

	fmt.Printf("Modified: %s\n", entry.Timestamp.Format("2006-01-02 15:04:05"))

	printDelim(settings.Delim, settings.Color)

	// Reset
	fmt.Printf("%s", getColor("default"))

	return nil

}

// Print an entry to the console with minimal data
func printEntryMinimal(entry *Entry, delim bool) error {

	var err error
	var settings *Settings

	err, settings = getOrCreateLocalConfig(APP)

	if err != nil {
		fmt.Printf("Error parsing config - \"%s\"\n", err.Error())
		return err
	}

	fmt.Printf("%s", getColor(strings.ToLower(settings.Color)))
	if strings.HasPrefix(settings.BgColor, "bg") {
		fmt.Printf("%s", getColor(strings.ToLower(settings.BgColor)))
	}

	if delim {
		printDelim(settings.Delim, settings.Color)
	}

	fmt.Printf("Title: %s\n", entry.Title)
	fmt.Printf("User: %s\n", entry.User)
	fmt.Printf("URL: %s\n", entry.Url)
	fmt.Printf("Modified: %s\n", entry.Timestamp.Format("2006-01-02 15:04:05"))

	printDelim(settings.Delim, settings.Color)

	// Reset
	fmt.Printf("%s", getColor("default"))

	return nil

}

// Read user input and return entered value
func readInput(reader *bufio.Reader, prompt string) string {

	var input string
	fmt.Printf(prompt + ": ")
	input, _ = reader.ReadString('\n')

	return strings.TrimSpace(input)
}

// Check for an active, decrypted database
func checkActiveDatabase() error {

	if !hasActiveDatabase() {
		fmt.Printf("No decrypted active database found.\n")
		return errors.New("no active database")
	}

	return nil
}

// Return true if active database is encrypted
func isActiveDatabaseEncrypted() bool {

	err, settings := getOrCreateLocalConfig(APP)
	if err == nil && settings.ActiveDB != "" {
		if _, err := os.Stat(settings.ActiveDB); err == nil {
			if _, flag := isFileEncrypted(settings.ActiveDB); flag {
				return true
			}
		}
	}

	return false
}

// Return true if always encrypt is on
func isEncryptOn() bool {

	_, settings := getOrCreateLocalConfig(APP)
	return settings.KeepEncrypted
}

// Combination of above 2 logic plus auto encryption on (a play on CryptOn)
func isActiveDatabaseEncryptedAndMaxKryptOn() (bool, string) {

	err, settings := getOrCreateLocalConfig(APP)
	if err == nil && settings.ActiveDB != "" {
		if _, err := os.Stat(settings.ActiveDB); err == nil {
			if _, flag := isFileEncrypted(settings.ActiveDB); flag && settings.KeepEncrypted && settings.AutoEncrypt {
				return true, settings.ActiveDB
			}
		}
	}

	return false, ""
}

// (Temporarily) enable showing of passwords
func setShowPasswords() error {
	//  fmt.Printf("Setting show passwords to true\n")
	settingsRider.ShowPasswords = true
	return nil
}

// Copy the password to clipboard - only for single listings or single search results
func setCopyPasswordToClipboard() error {
	settingsRider.CopyPassword = true
	return nil
}

func setAssumeYes() error {
	settingsRider.AssumeYes = true
	return nil
}

func setType(_type string) {
	settingsRider.Type = _type
}

func copyPasswordToClipboard(passwd string) {
	clipboard.WriteAll(passwd)
}

// Generate a random file name
func randomFileName(folder string, suffix string) string {

	_, name := generateRandomBytes(16)
	return filepath.Join(folder, hex.EncodeToString(name)+suffix)
}

// Detect card type from card number
func detectCardType(cardNum string) (string, error) {

	var cardTypeIndex creditcard.CardType
	var err error

	card := creditcard.Card{
		Type:        "N/A",
		Number:      cardNum,
		ExpiryMonth: 12,
		ExpiryYear:  99,
		CVV:         "999",
	}

	cardTypeIndex, err = card.DetermineCardType()
	if err != nil {
		return "", err
	}

	return creditcard.CardTypeNames[cardTypeIndex], nil
}

// Validate CVV
func validateCvv(cardCvv string, cardClass string) bool {

	var matched bool

	// Amex CVV is 4 digits, rest are 3
	if cardClass == "American Express" {
		if matched, _ = regexp.Match(`^\d{4}$`, []byte(cardCvv)); matched {
			return matched
		}
	} else {
		if matched, _ = regexp.Match(`^\d{3}$`, []byte(cardCvv)); matched {
			return matched
		}
	}

	return false
}

func validateCardPin(cardPin string) bool {

	// A PIN is 4 digits or more
	if matched, _ := regexp.Match(`^\d{4,}$`, []byte(cardPin)); matched {
		return matched
	}

	return false
}

// Verify if the expiry date is in the form mm/dd
func checkValidExpiry(expiryDate string) bool {

	pieces := strings.Split(expiryDate, "/")

	if len(pieces) == 2 {
		// Sofar, so good
		var month int
		var year int
		var err error

		month, err = strconv.Atoi(pieces[0])
		if err != nil {
			fmt.Printf("Error parsing month: %s: \"%s\"\n", month, err.Error())
			return false
		}
		year, err = strconv.Atoi(pieces[1])
		if err != nil {
			fmt.Printf("Error parsing year: %s: \"%s\"\n", year, err.Error())
			return false
		}

		// Month should be in range 1 -> 12
		if month < 1 || month > 12 {
			fmt.Printf("Error: invalid value for month - %d!\n", month)
			return false
		}
		// Year should be >= current year
		currYear, _ := strconv.Atoi(strconv.Itoa(time.Now().Year())[2:])
		if year < currYear {
			fmt.Printf("Error: year should be >= %d\n", currYear)
			return false
		}

		return true
	} else {
		fmt.Println("Error: invalid input")
		return false
	}

}
