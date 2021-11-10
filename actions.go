// Actions on the database
package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Print the current active database path
func showActiveDatabasePath() error {

	err, settings := getOrCreateLocalConfig(APP)

	if err != nil {
		fmt.Printf("Error parsing config - \"%s\"\n", err.Error())
		return err
	}

	if settings != nil {
		if settings.ActiveDB != "" {
			fmt.Printf("%s\n", settings.ActiveDB)
		} else {
			fmt.Println("No active database")
		}
		return nil
	} else {
		fmt.Printf("Error - null config\n")
		return errors.New("null config")
	}
}

// Set the current active database path
func setActiveDatabasePath(dbPath string) error {

	var fullPath string
	var activeEncrypted bool
	var newEncrypted bool

	err, settings := getOrCreateLocalConfig(APP)

	if err != nil {
		fmt.Printf("Error parsing config - \"%s\"\n", err.Error())
		return err
	}

	if settings != nil {
		var flag bool

		if _, err = os.Stat(dbPath); os.IsNotExist(err) {
			fmt.Printf("Error - path %s does not exist\n", dbPath)
			return err
		}

		fullPath, _ = filepath.Abs(dbPath)

		if fullPath == settings.ActiveDB {
			fmt.Printf("Current database is \"%s\" - nothing to do\n", fullPath)
			return nil
		}

		if _, flag = isFileEncrypted(settings.ActiveDB); flag {
			activeEncrypted = true
		}

		if _, flag = isFileEncrypted(fullPath); flag {
			newEncrypted = true
		}

		// If autoencrypt is true - encrypt current DB automatically
		if settings.AutoEncrypt {
			if !activeEncrypted {
				fmt.Printf("Encrypting current active database - %s\n", settings.ActiveDB)
				err = encryptActiveDatabase()
				if err == nil {
					activeEncrypted = true
				}
			}

			if newEncrypted {
				// Decrypt new database if it is encrypted
				fmt.Printf("Database %s is encrypted, decrypting it\n", fullPath)
				err, _ = decryptDatabase(fullPath)
				if err != nil {
					fmt.Printf("Decryption Error - \"%s\", not switching databases\n", err.Error())
					return err
				} else {
					newEncrypted = false
				}
			}
		}

		if !activeEncrypted {
			// Use should manually encrypt before switching
			fmt.Println("Auto-encrypt disabled, encrypt existing database before switching to new.")
			return nil
		}

		if newEncrypted {
			// Use should manually decrypt before switching
			fmt.Println("Auto-encrypt disabled, decrypt new database manually before switching.")
			return nil
		}

		settings.ActiveDB = fullPath
		err = updateSettings(settings, settings.ConfigPath)
		if err == nil {
			fmt.Println("Switched active database successfully.")
		} else {
			fmt.Printf("Error updating settings - \"%s\"\n", err.Error())
		}

		return err

	} else {
		fmt.Printf("Error - null config\n")
		return errors.New("null config")
	}
}

// Text menu driven function to add a new entry
func addNewEntry() error {

	var userName string
	var title string
	var url string
	var notes string
	var passwd string
	var err error
	var maxKrypt bool
	var encPasswd string
	var defaultDB string

	maxKrypt, defaultDB = isActiveDatabaseEncryptedAndMaxKryptOn()

	// If max krypt on - then autodecrypt on call and auto encrypt after call
	if maxKrypt {
		err, encPasswd = decryptDatabase(defaultDB)
		if err != nil {
			return err
		}
	}

	if err = checkActiveDatabase(); err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)
	title = readInput(reader, "Title")
	url = readInput(reader, "URL")

	if len(url) > 0 && !strings.HasPrefix(strings.ToLower(url), "http://") && !strings.HasPrefix(strings.ToLower(url), "https://") {
		url = "http://" + url
	}

	userName = readInput(reader, "Username")

	fmt.Printf("Password (enter to generate new): ")
	err, passwd = readPassword()

	if len(passwd) == 0 {
		fmt.Printf("\nGenerating password ...")
		err, passwd = generateRandomPassword(16)
		fmt.Printf("done")
	}
	//	fmt.Printf("Password => %s\n", passwd)

	notes = readInput(reader, "\nNotes")

	// Title and username/password are mandatory
	if len(title) == 0 {
		fmt.Printf("Error - valid Title required\n")
		return errors.New("invalid input")
	}
	if len(userName) == 0 {
		fmt.Printf("Error - valid Username required\n")
		return errors.New("invalid input")
	}
	if len(passwd) == 0 {
		fmt.Printf("Error - valid Password required\n")
		return errors.New("invalid input")
	}

	// Trim spaces
	err = addNewDatabaseEntry(title, userName, url, passwd, notes)

	if err != nil {
		fmt.Printf("Error adding entry - \"%s\"\n", err.Error())
	}

	// If max krypt on - then autodecrypt on call and auto encrypt after call
	if maxKrypt {
		err = encryptDatabase(defaultDB, &encPasswd)
	}

	return err
}

// Edit a current entry by id
func editCurrentEntry(idString string) error {

	var userName string
	var title string
	var url string
	var notes string
	var passwd string
	var err error
	var entry *Entry
	var maxKrypt bool
	var defaultDB string
	var encPasswd string
	var id int

	maxKrypt, defaultDB = isActiveDatabaseEncryptedAndMaxKryptOn()

	// If max krypt on - then autodecrypt on call and auto encrypt after call
	if maxKrypt {
		err, encPasswd = decryptDatabase(defaultDB)
		if err != nil {
			return err
		}
	}

	if err = checkActiveDatabase(); err != nil {
		return err
	}

	id, _ = strconv.Atoi(idString)

	err, entry = getEntryById(id)
	if err != nil || entry == nil {
		fmt.Printf("No entry found for id %d\n", id)
		return err
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Current Title: %s\n", entry.Title)
	title = readInput(reader, "New Title")

	fmt.Printf("Current URL: %s\n", entry.Url)
	url = readInput(reader, "New URL")

	if len(url) > 0 && !strings.HasPrefix(strings.ToLower(url), "http://") && !strings.HasPrefix(strings.ToLower(url), "https://") {
		url = "http://" + url
	}

	fmt.Printf("Current Username: %s\n", entry.User)
	userName = readInput(reader, "New Username")

	fmt.Printf("Current Password: %s\n", entry.Password)
	fmt.Printf("New Password ([y/Y] to generate new, enter will keep old one): ")
	err, passwd = readPassword()

	if strings.ToLower(passwd) == "y" {
		fmt.Printf("\nGenerating new password ...")
		err, passwd = generateRandomPassword(16)
	}
	//	fmt.Printf("Password => %s\n", passwd)

	fmt.Printf("\nCurrent Notes: %s\n", entry.Notes)
	notes = readInput(reader, "New Notes")

	// Update
	err = updateDatabaseEntry(entry, title, userName, url, passwd, notes)
	if err != nil {
		fmt.Printf("Error updating entry - \"%s\"\n", err.Error())
	}

	// If max krypt on - then autodecrypt on call and auto encrypt after call
	if maxKrypt {
		err = encryptDatabase(defaultDB, &encPasswd)
	}

	return err
}

// List current entry by id
func listCurrentEntry(idString string) error {

	var id int
	var err error
	var entry *Entry
	var maxKrypt bool
	var defaultDB string
	var passwd string

	maxKrypt, defaultDB = isActiveDatabaseEncryptedAndMaxKryptOn()

	// If max krypt on - then autodecrypt on call and auto encrypt after call
	if maxKrypt {
		err, passwd = decryptDatabase(defaultDB)
		if err != nil {
			return err
		}
	}

	if err = checkActiveDatabase(); err != nil {
		return err
	}

	id, _ = strconv.Atoi(idString)

	//	fmt.Printf("Listing current entry - %d\n", id)
	err, entry = getEntryById(id)
	if err != nil || entry == nil {
		fmt.Printf("No entry found for id %d\n", id)
		return err
	}

	err = printEntry(entry, true)

	// If max krypt on - then autodecrypt on call and auto encrypt after call
	if maxKrypt {
		err = encryptDatabase(defaultDB, &passwd)
	}

	return err
}

// List all entries
func listAllEntries() error {

	var err error
	var maxKrypt bool
	var defaultDB string
	var passwd string

	maxKrypt, defaultDB = isActiveDatabaseEncryptedAndMaxKryptOn()

	// If max krypt on - then autodecrypt on call and auto encrypt after call
	if maxKrypt {
		err, passwd = decryptDatabase(defaultDB)
		if err != nil {
			return err
		}
	}

	if err = checkActiveDatabase(); err != nil {
		return err
	}

	err, settings := getOrCreateLocalConfig(APP)

	if err != nil {
		fmt.Printf("Error parsing config - \"%s\"\n", err.Error())
		return err
	}

	orderKeys := strings.Split(settings.ListOrder, ",")
	err, entries := iterateEntries(orderKeys[0], orderKeys[1])

	if err == nil {
		if len(entries) > 0 {
			fmt.Println("=====================================================================")
			for _, entry := range entries {
				printEntry(&entry, false)
			}
		} else {
			fmt.Println("No entries.")
		}
	} else {
		fmt.Printf("Error fetching entries: \"%s\"\n", err.Error())
		return err
	}

	// If max krypt on - then autodecrypt on call and auto encrypt after call
	if maxKrypt {
		err = encryptDatabase(defaultDB, &passwd)
	}

	return err
}

// Find current entry by term - prints all matches
func findCurrentEntry(term string) error {

	var err error
	var entries []Entry
	var maxKrypt bool
	var defaultDB string
	var passwd string

	maxKrypt, defaultDB = isActiveDatabaseEncryptedAndMaxKryptOn()

	// If max krypt on - then autodecrypt on call and auto encrypt after call
	if maxKrypt {
		err, passwd = decryptDatabase(defaultDB)
		if err != nil {
			return err
		}
	}

	if err = checkActiveDatabase(); err != nil {
		return err
	}

	err, entries = searchDatabaseEntry(term)
	if err != nil || len(entries) == 0 {
		fmt.Printf("Entry for query \"%s\" not found\n", term)
		return err
	} else {
		var delim bool

		if len(entries) == 1 {
			delim = true
		} else {
			fmt.Println("=====================================================================")
		}

		for _, entry := range entries {
			printEntry(&entry, delim)
		}
	}

	// If max krypt on - then autodecrypt on call and auto encrypt after call
	if maxKrypt {
		err = encryptDatabase(defaultDB, &passwd)
	}

	return err
}

// Remove current entry by id
func removeCurrentEntry(idString string) error {

	var err error
	var entry *Entry
	var id int
	var maxKrypt bool
	var defaultDB string
	var passwd string

	maxKrypt, defaultDB = isActiveDatabaseEncryptedAndMaxKryptOn()

	// If max krypt on - then autodecrypt on call and auto encrypt after call
	if maxKrypt {
		err, passwd = decryptDatabase(defaultDB)
		if err != nil {
			return err
		}
	}

	if err = checkActiveDatabase(); err != nil {
		return err
	}

	id, _ = strconv.Atoi(idString)

	err, entry = getEntryById(id)
	if err != nil || entry == nil {
		fmt.Printf("No entry with id %d was found\n", id)
		return err
	}

	// Drop from the database
	err = removeDatabaseEntry(entry)
	if err == nil {
		fmt.Printf("Entry with id %d was removed from the database\n", id)
	}

	// If max krypt on - then autodecrypt on call and auto encrypt after call
	if maxKrypt {
		err = encryptDatabase(defaultDB, &passwd)
	}

	return err
}

// Copy current entry by id into new entry
func copyCurrentEntry(idString string) error {

	var err error
	var entry *Entry
	var id int
	var maxKrypt bool
	var defaultDB string
	var passwd string

	maxKrypt, defaultDB = isActiveDatabaseEncryptedAndMaxKryptOn()

	// If max krypt on - then autodecrypt on call and auto encrypt after call
	if maxKrypt {
		err, passwd = decryptDatabase(defaultDB)
		if err != nil {
			return err
		}
	}

	if err = checkActiveDatabase(); err != nil {
		return err
	}

	id, _ = strconv.Atoi(idString)

	err, entry = getEntryById(id)
	if err != nil || entry == nil {
		fmt.Printf("No entry with id %d was found\n", id)
		return err
	}

	err, _ = cloneEntry(entry)
	if err != nil {
		fmt.Printf("Error cloning entry: \"%s\"\n", err.Error())
		return err
	}

	// If max krypt on - then autodecrypt on call and auto encrypt after call
	if maxKrypt {
		err = encryptDatabase(defaultDB, &passwd)
	}

	return err
}

// Encrypt the active database
func encryptActiveDatabase() error {

	var err error
	var dbPath string

	if err = checkActiveDatabase(); err != nil {
		return err
	}

	err, dbPath = getActiveDatabase()
	if err != nil {
		fmt.Printf("Error getting active database path - \"%s\"\n", err.Error())
		return err
	}

	return encryptDatabase(dbPath, nil)
}

// Encrypt the database using AES
func encryptDatabase(dbPath string, givenPasswd *string) error {

	var err error
	var passwd string
	var passwd2 string

	// If password is given, use it
	if givenPasswd != nil {
		passwd = *givenPasswd
	}

	if len(passwd) == 0 {
		fmt.Printf("Password: ")
		err, passwd = readPassword()

		if err == nil {
			fmt.Printf("\nPassword again: ")
			err, passwd2 = readPassword()
			if err == nil {
				if passwd != passwd2 {
					fmt.Println("\nPassword mismatch.")
					return errors.New("mismatched passwords")
				}
			}
		}

		if err != nil {
			fmt.Printf("Error reading password - \"%s\"\n", err.Error())
			return err
		}
	}

	//	err = encryptFileAES(dbPath, passwd)
	_, settings := getOrCreateLocalConfig(APP)

	switch settings.Cipher {
	case "aes":
		err = encryptFileAES(dbPath, passwd)
	case "xchacha", "chacha", "xchachapoly":
		err = encryptFileXChachaPoly(dbPath, passwd)
	default:
		fmt.Println("No cipher set, defaulting to AES")
		err = encryptFileAES(dbPath, passwd)
	}

	if err == nil {
		fmt.Println("\nEncryption complete.")
	}

	return err
}

// Decrypt an encrypted database
func decryptDatabase(dbPath string) (error, string) {

	var err error
	var passwd string
	var flag bool

	if err, flag = isFileEncrypted(dbPath); !flag {
		fmt.Println(err.Error())
		return err, ""
	}

	fmt.Printf("Password: ")
	err, passwd = readPassword()

	if err != nil {
		fmt.Printf("\nError reading password - \"%s\"\n", err.Error())
		return err, ""
	}

	_, settings := getOrCreateLocalConfig(APP)

	switch settings.Cipher {
	case "aes":
		err = decryptFileAES(dbPath, passwd)
	case "xchacha", "chacha", "xchachapoly":
		err = decryptFileXChachaPoly(dbPath, passwd)
	default:
		fmt.Println("No cipher set, defaulting to AES")
		err = decryptFileAES(dbPath, passwd)
	}

	if err == nil {
		fmt.Println("\nDecryption complete.")
	}

	return err, passwd
}
