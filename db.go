// Database operations for varuh
package main

import (
	"database/sql"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Structure representing an entry in the db
type Entry struct {
	ID        int       `gorm:"column:id;autoIncrement;primaryKey"`
	Type      string    `gorm:"column:type"`  // Type of entry - password (default), card, identity etc
	Title     string    `gorm:"column:title"`
    Name      string    `gorm:"column:name"` // Card holder name/ID card name - for types cards/identity
	Company   string    `gorm:"column:company"` // Company name of person  - for type identity and
                                             	// Credit card company for type CC
	Number    string    `gorm:"column:number"`  // Number type - CC number for credit cards
	                                            // ID card number for identity types
	SecurityCode string  `gorm:"security_code"` // CVV number/security code for CC type
	ExpiryMonth string   `gorm:"expiry_month"` // CC or Identity document expiry month
	ExpiryDay string     `gorm:"expiry_day"`   // Identity document expiry day
	ExpiryYear string    `gorm:"expiry_year"`  // CC or Identity document expiry year
	FirstName  string    `gorm:"column:first_name"` // first name - for ID card types
	MiddleName string    `gorm:"column:middle_name"` // middle name - for ID card types
	LastName   string    `gorm:"column:last_name"` // last name - for ID card types	
	Email      string    `gorm:"email"` // Email - for ID card types
	PhoneNumber string   `gorm:"phone_number"` // Phone number - for ID card types

	Active    bool       `gorm:"active;default:true"` // Is the id card/CC active ?
	User      string    `gorm:"column:user"`
	Url       string    `gorm:"column:url"`
	Password  string    `gorm:"column:password"`
	Notes     string    `gorm:"column:notes"`
	Tags      string    `gorm:"column:tags"`
	Timestamp time.Time `gorm:"type:timestamp;default:(datetime('now','localtime'))"` // sqlite3
}

func (e *Entry) TableName() string {
	return "entries"
}

// Structure representing an extended entry in the db - for custom fields
type ExtendedEntry struct {
	ID         int       `gorm:"column:id;autoIncrement;primaryKey"`
	FieldName  string    `gorm:"column:field_name"`
	FieldValue string    `gorm:"column:field_value"`
	Timestamp  time.Time `gorm:"type:timestamp;default:(datetime('now','localtime'))"` // sqlite3

	Entry   Entry `gorm:"foreignKey:EntryID"`
	EntryID int
}

func (ex *ExtendedEntry) TableName() string {
	return "exentries"
}

type Address struct {
	ID     int   `gorm:"column:id;autoIncrement;primaryKey"`
	Number string `gorm:"column:number"` // Flat or building number
	Building string `gorm:"column:building"` // Apartment or building or society name
	Street  string `gorm:"column:street"` // Street address
	Locality string `gorm:"column:locality"` // Name of the locality e.g: Whitefield
	Area     string `gorm:"column:area"` // Name of the larger area e.g: East Bangalore
	City     string `gorm:"column:city"` // Name of the city e.g: Bangalore
	State   string  `gorm:"column:state"` // Name of the state e.g: Karnataka
	Country string  `gorm:"column:country"` // Name of the country e.g: India

	Landmark string `gorm:"column:landmark"` // Name of landmark if any
	ZipCode string  `gorm:"column:zipcode"` // PIN/ZIP code
	Type    string  `gorm:"column:type"` // Type of address: Home/Work/Business

	Entry   Entry `gorm:"foreignKey:EntryID"`
	EntryID int	
}

func (ad *Address) TableName() string {
	return "address"
}

// Clone an entry
func (e1 *Entry) Copy(e2 *Entry) {

	if e2 != nil {
		switch (e2.Type) {
		case "password":
			e1.Title = e2.Title
			e1.User = e2.User
			e1.Url = e2.Url
			e1.Password = e2.Password
			e1.Notes = e2.Notes
			e1.Tags = e2.Tags
			e1.Type = e2.Type
		case "card":
			e1.Title = e2.Title
			e1.Name = e2.Name // card holder name
			e1.Company = e2.Company
			e1.Number = e2.Number
			e1.SecurityCode = e2.SecurityCode
			e1.ExpiryMonth = e2.ExpiryMonth
			e1.ExpiryYear = e2.ExpiryYear
			e1.Tags = e2.Tags
			e1.Notes = e2.Notes
			e1.Type = e2.Type			
		case "identity":
			e1.Title = e2.Title
			e1.Name = e2.Name 
			e1.Company = e2.Company
			e1.FirstName = e2.FirstName
			e1.LastName = e2.LastName			
			e1.MiddleName = e2.MiddleName
			e1.User = e2.User
			e1.Email = e2.Email
			e1.PhoneNumber = e2.PhoneNumber
			e1.Number = e2.Number
			e1.Notes = e2.Notes
			e1.Tags = e2.Tags
			e1.Type = e2.Type
		}
	}
}

// Clone an entry
func (e1 *ExtendedEntry) Copy(e2 *ExtendedEntry) {

	if e2 != nil {
		e1.FieldName = e2.FieldName
		e1.FieldValue = e2.FieldValue
		e1.EntryID = e2.EntryID
	}
}

// Create a new database
func openDatabase(filePath string) (error, *gorm.DB) {

	db, err := gorm.Open(sqlite.Open(filePath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	return err, db
}

// Create a new table for Entries in the database
func createNewEntry(db *gorm.DB) error {
	return db.AutoMigrate(&Entry{})
}

// Create a new table for Extended Entries in the database
func createNewExEntry(db *gorm.DB) error {
	return db.AutoMigrate(&ExtendedEntry{})
}

// Init new database including tables
func initNewDatabase(dbPath string) error {

	var err error
	var db *gorm.DB
	var absPath string

	if hasActiveDatabase() {
		// Has an active database - encrypt it before creating new one
		_, activeDbPath := getActiveDatabase()
		absPath, _ = filepath.Abs(dbPath)

		if absPath == activeDbPath {
			fmt.Printf("Database already exists and is active - %s\n", dbPath)
			return nil
		} else {
			// TBD
			fmt.Printf("Encrytping current database - %s\n", activeDbPath)
			encryptDatabase(activeDbPath, nil)
		}
	}

	if _, err = os.Stat(dbPath); err == nil {
		// filePath exists, remove it
		os.Remove(dbPath)
	}

	err, db = openDatabase(dbPath)
	if err != nil {
		fmt.Printf("Error creating new database - \"%s\"\n", err.Error())
		return err
	}

	err = createNewEntry(db)
	if err != nil {
		fmt.Printf("Error creating schema - \"%s\"\n", err.Error())
		return err
	}

	err = createNewExEntry(db)
	if err != nil {
		fmt.Printf("Error creating schema - \"%s\"\n", err.Error())
		return err
	}

	fmt.Printf("Created new database - %s\n", dbPath)

	// Update config
	absPath, err = filepath.Abs(dbPath)
	// Chmod it
	os.Chmod(absPath, 0600)

	if err == nil {
		fmt.Printf("Updating active db path - %s\n", absPath)
		updateActiveDbPath(absPath)
	} else {
		fmt.Printf("Error - %s\n", err.Error())
		return err
	}

	return nil
}

// Open currently active database
func openActiveDatabase() (error, *gorm.DB) {

	var dbPath string
	var err error

	err, dbPath = getActiveDatabase()
	if err != nil {
		fmt.Printf("Error getting active database path - %s\n", err.Error())
		return err, nil
	}

	err, db := openDatabase(dbPath)
	if err != nil {
		fmt.Printf("Error opening active database path - %s: %s\n", dbPath, err.Error())
		return err, nil
	}

	return nil, db
}

// Add custom entries to a database entry
func addCustomEntries(db *gorm.DB, entry *Entry, customEntries []CustomEntry) error {

	var count int
	var err error

	err = createNewExEntry(db)
	if err != nil {
		fmt.Printf("Error creating schema - \"%s\"\n", err.Error())
		return err
	}

	for _, customEntry := range customEntries {
		var exEntry ExtendedEntry

		exEntry = ExtendedEntry{FieldName: customEntry.fieldName, FieldValue: customEntry.fieldValue,
			EntryID: entry.ID}

		resultEx := db.Create(&exEntry)
		if resultEx.Error == nil && resultEx.RowsAffected == 1 {
			count += 1
		}
	}

	fmt.Printf("Created %d custom entries for entry: %d.\n", count, entry.ID)
	return nil
}

// Replace custom entries to a database entry (Drop existing and add fresh)
func replaceCustomEntries(db *gorm.DB, entry *Entry, updatedEntries []CustomEntry) error {

	var count int
	var err error
	var customEntries []ExtendedEntry

	err = createNewExEntry(db)
	if err != nil {
		fmt.Printf("Error creating schema - \"%s\"\n", err.Error())
		return err
	}

	db.Where("entry_id = ?", entry.ID).Delete(&customEntries)

	for _, customEntry := range updatedEntries {
		var exEntry ExtendedEntry

		exEntry = ExtendedEntry{FieldName: customEntry.fieldName, FieldValue: customEntry.fieldValue,
			EntryID: entry.ID}

		resultEx := db.Create(&exEntry)
		if resultEx.Error == nil && resultEx.RowsAffected == 1 {
			count += 1
		}
	}

	fmt.Printf("Created %d custom entries for entry: %d.\n", count, entry.ID)
	return nil
}

// Add a new entry to current database
func addNewDatabaseEntry(title, userName, url, passwd, tags string,
	notes string, customEntries []CustomEntry) error {

	var entry Entry
	var err error
	var db *gorm.DB

	entry = Entry{Title: title, User: userName, Url: url, Password: passwd, Tags: strings.TrimSpace(tags),
		Notes: notes}

	err, db = openActiveDatabase()
	if err == nil && db != nil {
		//      result := db.Debug().Create(&entry)
		result := db.Create(&entry)
		if result.Error == nil && result.RowsAffected == 1 {
			// Add custom fields if given
			fmt.Printf("Created new entry with id: %d.\n", entry.ID)
			if len(customEntries) > 0 {
				return addCustomEntries(db, &entry, customEntries)
			}
			return nil
		} else if result.Error != nil {
			return result.Error
		}
	}

	return err
}

// Update current database entry with new values
func updateDatabaseEntry(entry *Entry, title, userName, url, passwd, tags string,
	notes string, customEntries []CustomEntry, flag bool) error {

	var updateMap map[string]interface{}

	updateMap = make(map[string]interface{})

	keyValMap := map[string]string{
		"title":    title,
		"user":     userName,
		"url":      url,
		"password": passwd,
		"notes":    notes,
		"tags":     tags}

	for key, val := range keyValMap {
		if len(val) > 0 {
			updateMap[key] = val
		}
	}

	if len(updateMap) == 0 && !flag {
		fmt.Printf("Nothing to update\n")
		return nil
	}

	// Update timestamp also
	updateMap["timestamp"] = time.Now()

	err, db := openActiveDatabase()

	if err == nil && db != nil {
		result := db.Model(entry).Updates(updateMap)
		if result.Error != nil {
			return result.Error
		}

		if flag {
			replaceCustomEntries(db, entry, customEntries)
		}
		fmt.Println("Updated entry.")
		return nil
	}

	return err
}

// Find entry given the id
func getEntryById(id int) (error, *Entry) {

	var entry Entry
	var err error
	var db *gorm.DB

	err, db = openActiveDatabase()
	if err == nil && db != nil {
		result := db.First(&entry, id)
		if result.Error == nil {
			return nil, &entry
		} else {
			return result.Error, nil
		}
	}

	return err, nil
}

// Search database for the given string and return all matches
func searchDatabaseEntry(term string) (error, []Entry) {

	var entries []Entry
	var err error
	var db *gorm.DB
	var searchTerm string

	err, db = openActiveDatabase()
	if err == nil && db != nil {
		var conditions []string
		var condition string

		searchTerm = fmt.Sprintf("%%%s%%", term)
		// Search on fields title, user, url and notes
		for _, field := range []string{"title", "user", "url", "notes"} {
			conditions = append(conditions, field+" like ?")
		}

		condition = strings.Join(conditions, " OR ")
		query := db.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm)
		res := query.Find(&entries)

		if res.Error != nil {
			return res.Error, nil
		}

		return nil, entries
	}

	return err, entries

}

// Union of two entry arrays
func union(entry1 []Entry, entry2 []Entry) []Entry {

	m := make(map[int]bool)

	for _, item := range entry1 {
		m[item.ID] = true
	}

	for _, item := range entry2 {
		if _, ok := m[item.ID]; !ok {
			entry1 = append(entry1, item)
		}
	}

	return entry1
}

// Intersection of two entry arrays
func intersection(entry1 []Entry, entry2 []Entry) []Entry {

	var common []Entry

	m := make(map[int]bool)

	for _, item := range entry1 {
		m[item.ID] = true
	}

	for _, item := range entry2 {
		if _, ok := m[item.ID]; ok {
			common = append(common, item)
		}
	}

	return common
}

// Search database for the given terms and returns matches according to operator
func searchDatabaseEntries(terms []string, operator string) (error, []Entry) {

	var err error
	var finalEntries []Entry

	for idx, term := range terms {
		var entries []Entry

		err, entries = searchDatabaseEntry(term)
		if err != nil {
			fmt.Printf("Error searching for term: %s - \"%s\"\n", term, err.Error())
			return err, entries
		}

		if idx == 0 {
			finalEntries = entries
		} else {
			if operator == "AND" {
				finalEntries = intersection(finalEntries, entries)
			} else if operator == "OR" {
				finalEntries = union(finalEntries, entries)
			}
		}
	}

	return nil, finalEntries
}

// Remove a given database entry
func removeDatabaseEntry(entry *Entry) error {

	var err error
	var db *gorm.DB

	err, db = openActiveDatabase()
	if err == nil && db != nil {
		var exEntries []ExtendedEntry

		res := db.Delete(entry)
		if res.Error != nil {
			return res.Error
		}

		// Delete extended entries if any
		exEntries = getExtendedEntries(entry)
		if len(exEntries) > 0 {
			res = db.Delete(exEntries)
			if res.Error != nil {
				return res.Error
			}
		}

		return nil
	}

	return err
}

// Clone an entry and return cloned entry
func cloneEntry(entry *Entry) (error, *Entry) {

	var entryNew Entry
	var err error
	var db *gorm.DB

	err, db = openActiveDatabase()
	if err == nil && db != nil {
		entryNew.Copy(entry)

		result := db.Create(&entryNew)
		if result.Error == nil && result.RowsAffected == 1 {
			fmt.Printf("Cloned to new entry, id: %d.\n", entryNew.ID)
			return nil, &entryNew
		} else if result.Error != nil {
			return result.Error, nil
		}
	}

	return err, nil
}

// Clone extended entries for an entry and return error code
func cloneExtendedEntries(entry *Entry, exEntries []ExtendedEntry) error {

	var err error
	var db *gorm.DB

	err, db = openActiveDatabase()
	if err == nil && db != nil {
		for _, exEntry := range exEntries {
			var exEntryNew ExtendedEntry

			exEntryNew.Copy(&exEntry)
			// Update the ID!
			exEntryNew.EntryID = entry.ID

			result := db.Create(&exEntryNew)
			if result.Error != nil {
				return result.Error
			}
		}
	}

	return err
}

// Return an iterator over all entries using the given order query keys
func iterateEntries(orderKey string, order string) (error, []Entry) {

	var err error
	var db *gorm.DB
	var entries []Entry

	err, db = openActiveDatabase()

	if err == nil && db != nil {
		var rows *sql.Rows

		rows, err = db.Model(&Entry{}).Order(fmt.Sprintf("%s %s", orderKey, order)).Rows()
		for rows.Next() {
			var entry Entry

			db.ScanRows(rows, &entry)
			entries = append(entries, entry)
		}

		return nil, entries
	}

	return err, nil
}

// Export all entries to string array
func entriesToStringArray(skipLongFields bool) (error, [][]string) {

	var err error
	var db *gorm.DB
	var dataArray [][]string

	err, db = openActiveDatabase()

	if err == nil && db != nil {
		var rows *sql.Rows
		var count int64

		db.Model(&Entry{}).Count(&count)

		dataArray = make([][]string, 0, count)

		rows, err = db.Model(&Entry{}).Order("id asc").Rows()
		for rows.Next() {
			var entry Entry
			var entryData []string

			db.ScanRows(rows, &entry)

			if skipLongFields {
				// Skip Notes
				entryData = []string{strconv.Itoa(entry.ID), entry.Title, entry.User, entry.Password, entry.Timestamp.Format("2006-06-02 15:04:05")}
			} else {
				entryData = []string{strconv.Itoa(entry.ID), entry.Title, entry.User, entry.Url, entry.Password, entry.Notes, entry.Timestamp.Format("2006-06-02 15:04:05")}
			}

			dataArray = append(dataArray, entryData)
		}
	}

	return err, dataArray
}

// Get extended entries associated to an entry
func getExtendedEntries(entry *Entry) []ExtendedEntry {

	var err error
	var db *gorm.DB
	var customEntries []ExtendedEntry

	err, db = openActiveDatabase()

	if err == nil && db != nil {
		db.Where("entry_id = ?", entry.ID).Find(&customEntries)
	}

	return customEntries
}
