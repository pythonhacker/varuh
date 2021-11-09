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
	"strings"
	"time"
)

// Structure representing an entry in the db
type Entry struct {
	ID        int       `gorm:"column:id;autoIncrement;primaryKey"`
	Title     string    `gorm:"column:title"`
	User      string    `gorm:"column:user"`
	Url       string    `gorm:"column:url"`
	Password  string    `gorm:"column:password"`
	Notes     string    `gorm:"column:notes"`
	Timestamp time.Time `gorm:"type:timestamp;default:(datetime('now','localtime'))"` // sqlite3
}

func (e *Entry) TableName() string {
	return "entries"
}

// Clone an entry
func (e1 *Entry) Copy(e2 *Entry) {

	if e2 != nil {
		e1.Title = e2.Title
		e1.User = e2.User
		e1.Url = e2.Url
		e1.Password = e2.Password
		e1.Notes = e2.Notes
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
			encryptDatabase(activeDbPath)
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

	fmt.Printf("Created new database - %s\n", dbPath)

	// Update config
	absPath, err = filepath.Abs(dbPath)

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

// Add a new entry to current database
func addNewDatabaseEntry(title, userName, url, passwd, notes string) error {

	var entry Entry
	var err error
	var db *gorm.DB

	entry = Entry{Title: title, User: userName, Url: url, Password: passwd, Notes: notes}

	err, db = openActiveDatabase()
	if err == nil && db != nil {
		//		result := db.Debug().Create(&entry)
		result := db.Create(&entry)
		if result.Error == nil && result.RowsAffected == 1 {
			fmt.Printf("Created new entry with id: %d\n.", entry.ID)
			return nil
		} else if result.Error != nil {
			return result.Error
		}
	}

	return err
}

// Update current database entry with new values
func updateDatabaseEntry(entry *Entry, title, userName, url, passwd, notes string) error {

	var updateMap map[string]interface{}

	updateMap = make(map[string]interface{})

	keyValMap := map[string]string{"title": title, "user": userName, "url": url, "password": passwd, "notes": notes}

	for key, val := range keyValMap {
		if len(val) > 0 {
			updateMap[key] = val
		}
	}

	if len(updateMap) == 0 {
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

// Remove a given database entry
func removeDatabaseEntry(entry *Entry) error {

	var err error
	var db *gorm.DB

	err, db = openActiveDatabase()
	if err == nil && db != nil {
		res := db.Delete(entry)
		if res.Error != nil {
			return res.Error
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
			fmt.Printf("Cloned to new entry, id: %d\n.", entryNew.ID)
			return nil, &entryNew
		} else if result.Error != nil {
			return result.Error, nil
		}
	}

	return err, nil
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
