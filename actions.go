// Actions on the database
package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

type CustomEntry struct {
	fieldName  string
	fieldValue string
}

// Wrappers (closures) for functions accepting strings as input for in/out encryption
func WrapperMaxKryptStringFunc(fn actionFunc) actionFunc {

	return func(inputStr string) error {
		var maxKrypt bool
		var defaultDB string
		var encPasswd string
		var err error

		maxKrypt, defaultDB = isActiveDatabaseEncryptedAndMaxKryptOn()

		// If max krypt on - then autodecrypt on call and auto encrypt after call
		if maxKrypt {
			err, encPasswd = decryptDatabase(defaultDB)
			if err != nil {
				return err
			}

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

			go func() {
				sig := <-sigChan
				fmt.Println("Received signal", sig)
				// Reencrypt
				encryptDatabase(defaultDB, &encPasswd)
				os.Exit(1)
			}()
		}

		err = fn(inputStr)

		// If max krypt on - then autodecrypt on call and auto encrypt after call
		if maxKrypt {
			encryptDatabase(defaultDB, &encPasswd)
		}

		return err
	}

}

// Wrappers (closures) for functions accepting no input for in/out encryption
func WrapperMaxKryptVoidFunc(fn voidFunc) voidFunc {

	return func() error {
		var maxKrypt bool
		var defaultDB string
		var encPasswd string
		var err error

		maxKrypt, defaultDB = isActiveDatabaseEncryptedAndMaxKryptOn()

		// If max krypt on - then autodecrypt on call and auto encrypt after call
		if maxKrypt {
			err, encPasswd = decryptDatabase(defaultDB)
			if err != nil {
				return err
			}

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

			go func() {
				sig := <-sigChan
				fmt.Println("Received signal", sig)
				// Reencrypt
				encryptDatabase(defaultDB, &encPasswd)
				os.Exit(1)
			}()
		}

		err = fn()

		// If max krypt on - then autodecrypt on call and auto encrypt after call
		if maxKrypt {
			encryptDatabase(defaultDB, &encPasswd)
		}

		return err
	}

}

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
				if !settings.AutoEncrypt {
					// Decrypt new database if it is encrypted
					fmt.Printf("Database %s is encrypted, decrypting it\n", fullPath)
					err, _ = decryptDatabase(fullPath)
					if err != nil {
						fmt.Printf("Decryption Error - \"%s\", not switching databases\n", err.Error())
						return err
					} else {
						newEncrypted = false
					}
				} else {
					// New database is encrypted and autoencrypt is set - so keep it like that
					// fmt.Printf("Database %s is already encrypted, nothing to do\n", fullPath)
				}
			}
		}

		if !activeEncrypted {
			// Use should manually encrypt before switching
			fmt.Println("Auto-encrypt disabled, encrypt existing database before switching to new.")
			return nil
		}

		if newEncrypted && !settings.AutoEncrypt {
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
	var tags string
	var err error
	var customEntries []CustomEntry

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
		err, passwd = generateStrongPassword()
		fmt.Printf("done")
	}
	//  fmt.Printf("Password => %s\n", passwd)

	tags = readInput(reader, "\nTags (separated by space): ")
	notes = readInput(reader, "Notes")

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

	customEntries = addCustomFields(reader)

	// Trim spaces
	err = addNewDatabaseEntry(title, userName, url, passwd, tags, notes, customEntries)

	if err != nil {
		fmt.Printf("Error adding entry - \"%s\"\n", err.Error())
	}

	return err
}

// Function to update existing custom entries and add new ones
// The bool part of the return value indicates whether to take action
func addOrUpdateCustomFields(reader *bufio.Reader, entry *Entry) ([]CustomEntry, bool) {

	var customEntries []ExtendedEntry
	var editedCustomEntries []CustomEntry
	var newCustomEntries []CustomEntry
	var flag bool

	customEntries = getExtendedEntries(entry)

	if len(customEntries) > 0 {

		fmt.Println("Editing/deleting custom fields")
		for _, customEntry := range customEntries {
			var fieldName string
			var fieldValue string

			fmt.Println("Field Name: " + customEntry.FieldName)
			fieldName = readInput(reader, "\tNew Field Name (Enter to keep, \"x\" to delete)")
			if strings.ToLower(strings.TrimSpace(fieldName)) == "x" {
				fmt.Println("Deleting field: " + customEntry.FieldName)
			} else {
				if strings.TrimSpace(fieldName) == "" {
					fieldName = customEntry.FieldName
				}

				fmt.Println("Field Value: " + customEntry.FieldValue)
				fieldValue = readInput(reader, "\tNew Field Value (Enter to keep)")
				if strings.TrimSpace(fieldValue) == "" {
					fieldValue = customEntry.FieldValue
				}

				editedCustomEntries = append(editedCustomEntries, CustomEntry{fieldName, fieldValue})
			}
		}
	}

	newCustomEntries = addCustomFields(reader)

	editedCustomEntries = append(editedCustomEntries, newCustomEntries...)

	// Cases where length == 0
	// 1. Existing entries - all deleted
	flag = len(customEntries) > 0 || len(editedCustomEntries) > 0

	return editedCustomEntries, flag
}

// Function to add custom fields to an entry
func addCustomFields(reader *bufio.Reader) []CustomEntry {

	// Custom fields
	var custom string
	var customEntries []CustomEntry

	custom = readInput(reader, "Do you want to add custom fields [y/N]")
	if strings.ToLower(custom) == "y" {

		fmt.Println("Keep entering custom field name followed by the value. Press return with no input once done.")
		for true {
			var customFieldName string
			var customFieldValue string

			customFieldName = strings.TrimSpace(readInput(reader, "Field Name"))
			if customFieldName != "" {
				customFieldValue = strings.TrimSpace(readInput(reader, "Value for "+customFieldName))
			}

			if customFieldName == "" && customFieldValue == "" {
				break
			}

			customEntries = append(customEntries, CustomEntry{customFieldName, customFieldValue})
		}
	}

	return customEntries
}

// Edit a current entry by id
func editCurrentEntry(idString string) error {

	var userName string
	var title string
	var url string
	var notes string
	var tags string
	var passwd string
	var err error
	var entry *Entry
	var id int

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
		err, passwd = generateStrongPassword()
	}
	//  fmt.Printf("Password => %s\n", passwd)

	fmt.Printf("\nCurrent Tags: %s\n", entry.Tags)
	tags = readInput(reader, "New Tags")

	fmt.Printf("\nCurrent Notes: %s\n", entry.Notes)
	notes = readInput(reader, "New Notes")

	customEntries, flag := addOrUpdateCustomFields(reader, entry)

	// Update
	err = updateDatabaseEntry(entry, title, userName, url, passwd, tags, notes, customEntries, flag)
	if err != nil {
		fmt.Printf("Error updating entry - \"%s\"\n", err.Error())
	}

	return err
}

// List current entry by id
func listCurrentEntry(idString string) error {

	var id int
	var err error
	var entry *Entry

	if err = checkActiveDatabase(); err != nil {
		return err
	}

	id, _ = strconv.Atoi(idString)

	//  fmt.Printf("Listing current entry - %d\n", id)
	err, entry = getEntryById(id)
	if err != nil || entry == nil {
		fmt.Printf("No entry found for id %d\n", id)
		return err
	}

	err = printEntry(entry, true)

	if err == nil && settingsRider.CopyPassword {
		//      fmt.Printf("Copying password " + entry.Password + " to clipboard\n")
		copyPasswordToClipboard(entry.Password)
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
			fmt.Printf("%s", getColor(strings.ToLower(settings.Color)))
			printDelim(settings.Delim, settings.Color)
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
	var terms []string

	if err = checkActiveDatabase(); err != nil {
		return err
	}

	terms = strings.Split(term, " ")

	err, entries = searchDatabaseEntries(terms, "AND")
	if err != nil || len(entries) == 0 {
		fmt.Printf("Entry for query \"%s\" not found\n", term)
		return err
	} else {
		var delim bool
		var pcopy bool

		if len(entries) == 1 {
			delim = true
			pcopy = true
			// Single entry means copy password can be enabled
		} else {
			_, settings := getOrCreateLocalConfig(APP)
			fmt.Printf("%s", getColor(strings.ToLower(settings.Color)))
			printDelim(settings.Delim, settings.Color)
		}

		for _, entry := range entries {
			printEntry(&entry, delim)
		}

		if pcopy && settingsRider.CopyPassword {
			// Single entry
			copyPasswordToClipboard(entries[0].Password)
		}
	}

	return err
}

// Remove a range of entries <id1>-<id2> say 10-14
func removeMultipleEntries(idRangeEntry string) error {

	var err error
	var idRange []string
	var id1, id2 int

	idRange = strings.Split(idRangeEntry, "-")

	if len(idRange) != 2 {
		fmt.Println("Invalid id range - " + idRangeEntry)
		return errors.New("Invalid id range - " + idRangeEntry)
	}

	id1, _ = strconv.Atoi(idRange[0])
	id2, _ = strconv.Atoi(idRange[1])

	if id1 >= id2 {
		fmt.Println("Invalid id range - " + idRangeEntry)
		return errors.New("Invalid id range - " + idRangeEntry)
	}

	for idNum := id1; idNum <= id2; idNum++ {
		err = removeCurrentEntry(fmt.Sprintf("%d", idNum))
	}

	return err
}

// Remove current entry by id
func removeCurrentEntry(idString string) error {

	var err error
	var entry *Entry
	var id int
	var response string

	if err = checkActiveDatabase(); err != nil {
		return err
	}

	if strings.Contains(idString, "-") {
		return removeMultipleEntries(idString)
	}

	id, _ = strconv.Atoi(idString)

	err, entry = getEntryById(id)
	if err != nil || entry == nil {
		fmt.Printf("No entry with id %d was found\n", id)
		return err
	}

	printEntryMinimal(entry, true)

	if !settingsRider.AssumeYes {
		response = readInput(bufio.NewReader(os.Stdin), "Please confirm removal [Y/n]: ")
	} else {
		response = "y"
	}

	if strings.ToLower(response) != "n" {
		// Drop from the database
		err = removeDatabaseEntry(entry)
		if err == nil {
			fmt.Printf("Entry with id %d was removed from the database\n", id)
		}
	} else {
		fmt.Println("Removal of entry cancelled by user.")
	}

	return err
}

// Copy current entry by id into new entry
func copyCurrentEntry(idString string) error {

	var err error
	var entry *Entry
	var entryNew *Entry
	var exEntries []ExtendedEntry

	var id int

	if err = checkActiveDatabase(); err != nil {
		return err
	}

	id, _ = strconv.Atoi(idString)

	err, entry = getEntryById(id)
	if err != nil || entry == nil {
		fmt.Printf("No entry with id %d was found\n", id)
		return err
	}

	err, entryNew = cloneEntry(entry)
	if err != nil {
		fmt.Printf("Error cloning entry: \"%s\"\n", err.Error())
		return err
	}

	exEntries = getExtendedEntries(entry)

	if len(exEntries) > 0 {
		fmt.Printf("%d extended entries found\n", len(exEntries))

		err = cloneExtendedEntries(entryNew, exEntries)
		if err != nil {
			fmt.Printf("Error cloning extended entries: \"%s\"\n", err.Error())
			return err
		}
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
		fmt.Printf("Encryption Password: ")
		err, passwd = readPassword()

		if err == nil {
			fmt.Printf("\nEncryption Password again: ")
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

	//  err = encryptFileAES(dbPath, passwd)
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

	fmt.Printf("Decryption Password: ")
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
		fmt.Println("...decryption complete.")
	}

	return err, passwd
}

// Migrate an existing database to the new schema
func migrateDatabase(dbPath string) error {

	var err error
	var flag bool
	var passwd string
	var db *gorm.DB

	if _, err = os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Printf("Error - path %s does not exist\n", dbPath)
		return err
	}

	if err, flag = isFileEncrypted(dbPath); flag {
		err, passwd = decryptDatabase(dbPath)
	}

	if err != nil {
		return err
	}

	err, db = openDatabase(dbPath)

	if err != nil {
		fmt.Printf("Error opening database path - %s: %s\n", dbPath, err.Error())
		return err
	}

	fmt.Println("Migrating tables ...")
	err = db.AutoMigrate(&Entry{})

	if err != nil {
		fmt.Printf("Error migrating table \"entries\" - %s: %s\n", dbPath, err.Error())
		return err
	}

	err = db.AutoMigrate(&ExtendedEntry{})

	if err != nil {
		fmt.Printf("Error migrating table \"exentries\" - %s: %s\n", dbPath, err.Error())
		return err
	}

	if flag {
		// File was encrypted - encrypt it again
		encryptDatabase(dbPath, &passwd)
	}

	fmt.Println("Migration successful.")

	return nil
}

// Export data to a varity of file types
func exportToFile(fileName string) error {

	var err error
	var maxKrypt bool
	var defaultDB string
	var passwd string

	ext := strings.ToLower(filepath.Ext(fileName))

	maxKrypt, defaultDB = isActiveDatabaseEncryptedAndMaxKryptOn()

	if ext == ".csv" || ext == ".md" || ext == ".html" || ext == ".pdf" {
		// If max krypt on - then autodecrypt on call and auto encrypt after call
		if maxKrypt {
			err, passwd = decryptDatabase(defaultDB)
			if err != nil {
				return err
			}
		}
	}

	switch ext {
	case ".csv":
		err = exportToCsv(fileName)
	case ".md":
		err = exportToMarkdown(fileName)
	case ".html":
		err = exportToHTML(fileName)
	case ".pdf":
		err = exportToPDF(fileName)
	default:
		fmt.Printf("Error - extn %s not supported\n", ext)
		return fmt.Errorf("format %s not supported", ext)
	}

	if err != nil {
		fmt.Printf("Error exporting to \"%s\" - \"%s\"\n", fileName, err.Error())
		return err
	} else {
		if _, err = os.Stat(fileName); err == nil {
			fmt.Printf("Exported to %s.\n", fileName)
			// Chmod 600
			os.Chmod(fileName, 0600)

			// If max krypt on - then autodecrypt on call and auto encrypt after call
			if maxKrypt {
				err = encryptDatabase(defaultDB, &passwd)
			}

			return err
		}
	}

	return err
}

// Export current database to markdown
func exportToMarkdown(fileName string) error {

	var err error
	var dataArray [][]string
	var fh *os.File
	var maxLengths [7]int
	var headers []string = []string{" ID ", " Title ", " User ", " URL ", " Password ", " Notes ", " Modified "}

	err, dataArray = entriesToStringArray(false)

	if err != nil {
		fmt.Printf("Error exporting entries to string array - \"%s\"\n", err.Error())
		return err
	}

	for _, record := range dataArray {
		for idx, field := range record {

			if len(field) > maxLengths[idx] {
				maxLengths[idx] = len(field)
			}
		}
	}

	//  fmt.Printf("%+v\n", maxLengths)
	fh, err = os.Create(fileName)
	if err != nil {
		fmt.Printf("Cannt open \"%s\" for writing - \"%s\"\n", fileName, err.Error())
		return err
	}

	defer fh.Close()

	writer := bufio.NewWriter(fh)

	// Write markdown header
	for idx, length := range maxLengths {
		delta := length - len(headers[idx])
		//      fmt.Printf("%d\n", delta)
		if delta > 0 {
			for i := 0; i < delta+2; i++ {
				headers[idx] += " "
			}
		}
	}

	writer.WriteString(" |" + strings.Join(headers, "|") + "|\n")

	// Write line separator
	writer.WriteString(" | ")
	for _, length := range maxLengths {

		for i := 0; i < length; i++ {
			writer.WriteString("-")
		}
		writer.WriteString(" | ")
	}
	writer.WriteString("\n")

	// Write records
	for _, record := range dataArray {
		writer.WriteString(" | ")
		for _, field := range record {
			writer.WriteString(field + " | ")
		}
		writer.WriteString("\n")
	}

	writer.Flush()

	return nil

}

// This needs pandoc and pdflatex support
func exportToPDF(fileName string) error {

	var err error
	var tmpFile string
	var passwd string
	var pdfTkFound bool

	cmd := exec.Command("which", "pandoc")
	if _, err = cmd.Output(); err != nil {
		return errors.New("pandoc not found")
	}

	cmd = exec.Command("which", "pdftk")
	if _, err = cmd.Output(); err != nil {
		fmt.Printf("pdftk not found, PDF won't be secure!\n")
	} else {
		pdfTkFound = true
	}

	if pdfTkFound {
		fmt.Printf("PDF Encryption Password: ")
		err, passwd = readPassword()
	}

	tmpFile = randomFileName(os.TempDir(), ".tmp")
	//  fmt.Printf("Temp file => %s\n", tmpFile)
	err = exportToMarkdownLimited(tmpFile)

	if err == nil {
		var args []string = []string{"-o", fileName, "-f", "markdown", "-V", "geometry:landscape", "--columns=600", "--pdf-engine", "xelatex", "--dpi=150", tmpFile}

		cmd = exec.Command("pandoc", args...)
		_, err = cmd.Output()
		// Remove tmpfile
		os.Remove(tmpFile)

		// If the file is generated, encrypt it if pdfTkFound
		if _, err = os.Stat(fileName); err == nil {
			fmt.Printf("\nFile %s created without password.\n", fileName)

			if pdfTkFound && len(passwd) > 0 {
				tmpFile = randomFileName(".", ".pdf")
				//              fmt.Printf("pdf file => %s\n", tmpFile)
				args = []string{fileName, "output", tmpFile, "user_pw", passwd}
				cmd = exec.Command("pdftk", args...)
				_, err = cmd.Output()

				if err == nil {
					// Copy over
					fmt.Printf("Added password to %s.\n", fileName)
					os.Remove(fileName)
					err = os.Rename(tmpFile, fileName)
				} else {
					fmt.Printf("Error adding password to pdf - \"%s\"\n", err.Error())
				}
			}
		}
	}

	return err

}

// Export current database to markdown minus the long fields
func exportToMarkdownLimited(fileName string) error {

	var err error
	var dataArray [][]string
	var fh *os.File
	var maxLengths [5]int
	var headers []string = []string{" ID ", " Title ", " User ", " Password ", " Modified "}

	err, dataArray = entriesToStringArray(true)

	if err != nil {
		fmt.Printf("Error exporting entries to string array - \"%s\"\n", err.Error())
		return err
	}

	for _, record := range dataArray {
		for idx, field := range record {

			if len(field) > maxLengths[idx] {
				maxLengths[idx] = len(field)
			}
		}
	}

	//  fmt.Printf("%+v\n", maxLengths)
	fh, err = os.Create(fileName)
	if err != nil {
		fmt.Printf("Cannt open \"%s\" for writing - \"%s\"\n", fileName, err.Error())
		return err
	}

	defer fh.Close()

	writer := bufio.NewWriter(fh)

	// Write markdown header
	for idx, length := range maxLengths {
		delta := length - len(headers[idx])
		//      fmt.Printf("%d\n", delta)
		if delta > 0 {
			for i := 0; i < delta+2; i++ {
				headers[idx] += " "
			}
		}
	}

	writer.WriteString(" |" + strings.Join(headers, "|") + "|\n")

	// Write line separator
	writer.WriteString(" | ")
	for _, length := range maxLengths {

		for i := 0; i < length; i++ {
			writer.WriteString("-")
		}
		writer.WriteString(" | ")
	}
	writer.WriteString("\n")

	// Write records
	for _, record := range dataArray {
		writer.WriteString(" | ")
		for _, field := range record {
			writer.WriteString(field + " | ")
		}
		writer.WriteString("\n")
	}

	writer.Flush()

	return nil

}

// Export current database to html
func exportToHTML(fileName string) error {

	var err error
	var dataArray [][]string
	var fh *os.File
	var headers []string = []string{" ID ", " Title ", " User ", " URL ", " Password ", " Notes ", " Modified "}

	err, dataArray = entriesToStringArray(false)

	if err != nil {
		fmt.Printf("Error exporting entries to string array - \"%s\"\n", err.Error())
		return err
	}

	//  fmt.Printf("%+v\n", maxLengths)
	fh, err = os.Create(fileName)
	if err != nil {
		fmt.Printf("Cannt open \"%s\" for writing - \"%s\"\n", fileName, err.Error())
		return err
	}

	defer fh.Close()

	writer := bufio.NewWriter(fh)

	writer.WriteString("<html><body>\n")
	writer.WriteString("<table cellPadding=\"2\" cellSpacing=\"2\" border=\"1\">\n")
	writer.WriteString("<theader>\n")

	for _, h := range headers {
		writer.WriteString(fmt.Sprintf("<th>%s</th>", h))
	}
	writer.WriteString("</theader>\n")
	writer.WriteString("<tbody>\n")

	// Write records
	for _, record := range dataArray {
		writer.WriteString("<tr>")
		for _, field := range record {
			writer.WriteString(fmt.Sprintf("<td>%s</td>", field))
		}
		writer.WriteString("</tr>\n")
	}
	writer.WriteString("</tbody>\n")
	writer.WriteString("</table>\n")

	writer.WriteString("</body></html>\n")

	writer.Flush()

	return nil

}

// Export current database to CSV
func exportToCsv(fileName string) error {

	var err error
	var dataArray [][]string
	var fh *os.File

	err, dataArray = entriesToStringArray(false)

	if err != nil {
		fmt.Printf("Error exporting entries to string array - \"%s\"\n", err.Error())
		return err
	}

	fh, err = os.Create(fileName)
	if err != nil {
		fmt.Printf("Cannt open \"%s\" for writing - \"%s\"\n", fileName, err.Error())
		return err
	}

	writer := csv.NewWriter(fh)

	// Write header
	writer.Write([]string{"ID", "Title", "User", "URL", "Password", "Notes", "Modified"})

	for idx, record := range dataArray {
		if err = writer.Write(record); err != nil {
			fmt.Printf("Error writing record #%d to %s - \"%s\"\n", idx+1, fileName, err.Error())
			break
		}
	}

	writer.Flush()

	if err != nil {
		return err
	}

	os.Chmod(fileName, 0600)
	fmt.Printf("!WARNING: Passwords are stored in plain-text!\n")
	fmt.Printf("Exported %d records to %s .\n", len(dataArray), fileName)

	return nil
}
