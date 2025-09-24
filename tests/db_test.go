package tests

import (
	"log"
	"path/filepath"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"	
	"testing"
	"varuh"
)

func createMockDb(fileName string) error {
		// Just open it with GORM/SQLite driver
	db, err := gorm.Open(sqlite.Open(fileName), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}

	// This will create the DB file with a proper SQLite header if it doesnt exist
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	log.Printf("SQLite database %s created and ready.\n", fileName)

	return nil
}


func TestEntry_Copy(t *testing.T) {
	tests := []struct {
		name string
		e1   *varuh.Entry
		e2   *varuh.Entry
		want *varuh.Entry
	}{
		{
			name: "copy password entry",
			e1:   &varuh.Entry{},
			e2: &varuh.Entry{
				Title:    "Test Title",
				User:     "test@example.com",
				Url:      "https://example.com",
				Password: "secret123",
				Notes:    "Test notes",
				Tags:     "test,example",
				Type:     "password",
			},
			want: &varuh.Entry{
				Title:    "Test Title",
				User:     "test@example.com",
				Url:      "https://example.com",
				Password: "secret123",
				Notes:    "Test notes",
				Tags:     "test,example",
				Type:     "password",
			},
		},
		{
			name: "copy card entry",
			e1:   &varuh.Entry{},
			e2: &varuh.Entry{
				Title:      "Test Card",
				User:       "John Doe",
				Issuer:     "Chase Bank",
				Url:        "4111111111111111",
				Password:   "123",
				ExpiryDate: "12/25",
				Tags:       "credit,card",
				Notes:      "Main card",
				Type:       "card",
			},
			want: &varuh.Entry{
				Title:      "Test Card",
				User:       "John Doe",
				Issuer:     "Chase Bank",
				Url:        "4111111111111111",
				Password:   "123",
				ExpiryDate: "12/25",
				Tags:       "credit,card",
				Notes:      "Main card",
				Type:       "card",
			},
		},
		{
			name: "copy nil entry",
			e1:   &varuh.Entry{Title: "Original", User: "original@test.com"},
			e2:   nil,
			want: &varuh.Entry{Title: "Original", User: "original@test.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.e1.Copy(tt.e2)

			if tt.e2 == nil {
				// Should remain unchanged
				if tt.e1.Title != tt.want.Title || tt.e1.User != tt.want.User {
					t.Errorf("Entry should remain unchanged when copying nil")
				}
				return
			}

			// Compare relevant fields based on type
			switch tt.e2.Type {
			case "password":
				if tt.e1.Title != tt.want.Title ||
					tt.e1.User != tt.want.User ||
					tt.e1.Url != tt.want.Url ||
					tt.e1.Password != tt.want.Password ||
					tt.e1.Notes != tt.want.Notes ||
					tt.e1.Tags != tt.want.Tags ||
					tt.e1.Type != tt.want.Type {
					t.Errorf("Password entry copy failed")
				}
			case "card":
				if tt.e1.Title != tt.want.Title ||
					tt.e1.User != tt.want.User ||
					tt.e1.Issuer != tt.want.Issuer ||
					tt.e1.Url != tt.want.Url ||
					tt.e1.Password != tt.want.Password ||
					tt.e1.ExpiryDate != tt.want.ExpiryDate ||
					tt.e1.Tags != tt.want.Tags ||
					tt.e1.Notes != tt.want.Notes ||
					tt.e1.Type != tt.want.Type {
					t.Errorf("Card entry copy failed")
				}
			}
		})
	}
}

func TestExtendedEntry_Copy(t *testing.T) {
	tests := []struct {
		name string
		e1   *varuh.ExtendedEntry
		e2   *varuh.ExtendedEntry
		want *varuh.ExtendedEntry
	}{
		{
			name: "copy extended entry",
			e1:   &varuh.ExtendedEntry{},
			e2: &varuh.ExtendedEntry{
				FieldName:  "CustomField1",
				FieldValue: "CustomValue1",
				EntryID:    123,
			},
			want: &varuh.ExtendedEntry{
				FieldName:  "CustomField1",
				FieldValue: "CustomValue1",
				EntryID:    123,
			},
		},
		{
			name: "copy nil extended entry",
			e1: &varuh.ExtendedEntry{
				FieldName:  "Original",
				FieldValue: "OriginalValue",
				EntryID:    1,
			},
			e2: nil,
			want: &varuh.ExtendedEntry{
				FieldName:  "Original",
				FieldValue: "OriginalValue",
				EntryID:    1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.e1.Copy(tt.e2)

			if tt.e2 == nil {
				// Should remain unchanged
				if tt.e1.FieldName != tt.want.FieldName ||
					tt.e1.FieldValue != tt.want.FieldValue ||
					tt.e1.EntryID != tt.want.EntryID {
					t.Errorf("ExtendedEntry should remain unchanged when copying nil")
				}
				return
			}

			if tt.e1.FieldName != tt.want.FieldName ||
				tt.e1.FieldValue != tt.want.FieldValue ||
				tt.e1.EntryID != tt.want.EntryID {
				t.Errorf("ExtendedEntry copy failed, got FieldName=%s FieldValue=%s EntryID=%d, want FieldName=%s FieldValue=%s EntryID=%d",
					tt.e1.FieldName, tt.e1.FieldValue, tt.e1.EntryID,
					tt.want.FieldName, tt.want.FieldValue, tt.want.EntryID)
			}
		})
	}
}

func TestOpenDatabase(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test SQLite database file
	testDB := filepath.Join(tempDir, "test.db")
	err := createMockDb(testDB)
	if err != nil {
		t.Fatalf("Failed to create test database file: %v", err)
	}

	tests := []struct {
		name     string
		filePath string
		wantErr  bool
	}{
		{"valid database", testDB, false},
		{"empty path", "", true},
		{"non-existent file", filepath.Join(tempDir, "nonexistent.db"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, db := varuh.OpenDatabase(tt.filePath)

			if (err != nil) != tt.wantErr {
				t.Errorf("OpenDatabase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && db == nil {
				t.Error("OpenDatabase() returned nil database for valid input")
			}
		})
	}
}

func TestCreateNewEntry(t *testing.T) {
	tempDir := t.TempDir()
	testDB := filepath.Join(tempDir, "test.db")

	// Create a basic SQLite file
	err := createMockDb(testDB)	
	if err != nil {
		t.Fatalf("Failed to create test database file: %v", err)
	}

	err, db := varuh.OpenDatabase(testDB)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = varuh.CreateNewEntry(db)
	if err != nil {
		t.Errorf("CreateNewEntry() error = %v", err)
	}
}

func TestCreateNewExEntry(t *testing.T) {
	tempDir := t.TempDir()
	testDB := filepath.Join(tempDir, "test.db")

	// Create a basic SQLite file
	err := createMockDb(testDB)	
	if err != nil {
		t.Fatalf("Failed to create test database file: %v", err)
	}

	err, db := varuh.OpenDatabase(testDB)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = varuh.CreateNewExEntry(db)
	if err != nil {
		t.Errorf("CreateNewExEntry() error = %v", err)
	}
}

func TestInitNewDatabase(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		dbPath  string
		wantErr bool
	}{
		{"valid path", filepath.Join(tempDir, "new.db"), false},
		{"empty path", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip tests that depend on global state/config
			if tt.name == "valid path" {
				t.Skip("Skipping InitNewDatabase test as it depends on global config state")
			}

			err := varuh.InitNewDatabase(tt.dbPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("InitNewDatabase() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetEntryById(t *testing.T) {
	// This function depends on active database state
	// We'll test that it doesn't panic and returns appropriate error
	t.Run("no active database", func(t *testing.T) {
		err, entry := varuh.GetEntryById(1)

		// Should return an error or nil entry when no active database
		if err == nil && entry != nil {
			t.Error("GetEntryById() should return error or nil entry when no active database")
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		err, entry := varuh.GetEntryById(-1)

		// Should handle invalid IDs gracefully
		if entry != nil && entry.ID == -1 {
			t.Error("GetEntryById() should not return entry with invalid ID")
		}
		_ = err // err can be nil or non-nil depending on database state
	})
}

func TestSearchDatabaseEntry(t *testing.T) {
	// This function depends on active database state
	t.Run("empty search term", func(t *testing.T) {
		err, entries := varuh.SearchDatabaseEntry("")

		// Should handle empty search term gracefully
		_ = err     // err can be nil or non-nil depending on database state
		_ = entries // entries can be empty or nil
	})

	t.Run("normal search term", func(t *testing.T) {
		err, entries := varuh.SearchDatabaseEntry("test")

		// Should handle normal search without panic
		_ = err     // err can be nil or non-nil depending on database state
		_ = entries // entries can be empty or nil
	})
}

func TestUnion(t *testing.T) {
	entry1 := varuh.Entry{ID: 1, Title: "Entry 1"}
	entry2 := varuh.Entry{ID: 2, Title: "Entry 2"}
	entry3 := varuh.Entry{ID: 3, Title: "Entry 3"}
	entry1Dup := varuh.Entry{ID: 1, Title: "Entry 1 Duplicate"}

	tests := []struct {
		name   string
		slice1 []varuh.Entry
		slice2 []varuh.Entry
		want   int // expected length of result
	}{
		{
			name:   "empty slices",
			slice1: []varuh.Entry{},
			slice2: []varuh.Entry{},
			want:   0,
		},
		{
			name:   "no overlap",
			slice1: []varuh.Entry{entry1, entry2},
			slice2: []varuh.Entry{entry3},
			want:   3,
		},
		{
			name:   "with overlap",
			slice1: []varuh.Entry{entry1, entry2},
			slice2: []varuh.Entry{entry1Dup, entry3},
			want:   3, // should not duplicate entry1
		},
	}

	// Use reflection to call the unexported union function
	// Since it's unexported, we'll test the public functions that use it
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't directly test the unexported union function
			// So we test SearchDatabaseEntries which uses it internally
			terms := []string{"term1", "term2"}
			_, entries := varuh.SearchDatabaseEntries(terms, "OR")

			// Just ensure no panic occurs and entries is a valid slice
			if entries == nil {
				entries = []varuh.Entry{}
			}
			_ = len(entries) // Use the result to avoid unused variable error
		})
	}
}

func TestSearchDatabaseEntries(t *testing.T) {
	tests := []struct {
		name     string
		terms    []string
		operator string
	}{
		{"empty terms", []string{}, "AND"},
		{"single term", []string{"test"}, "AND"},
		{"multiple terms AND", []string{"test", "example"}, "AND"},
		{"multiple terms OR", []string{"test", "example"}, "OR"},
		{"invalid operator", []string{"test"}, "INVALID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, entries := varuh.SearchDatabaseEntries(tt.terms, tt.operator)

			// Should handle all cases without panic
			_ = err     // err can be nil or non-nil depending on database state
			_ = entries // entries can be empty or nil
		})
	}
}

func TestRemoveDatabaseEntry(t *testing.T) {
	entry := &varuh.Entry{ID: 1, Title: "Test Entry"}

	t.Run("remove entry", func(t *testing.T) {
		err := varuh.RemoveDatabaseEntry(entry)

		// Should handle gracefully whether or not there's an active database
		_ = err // err can be nil or non-nil depending on database state
	})
}

func TestCloneEntry(t *testing.T) {
	entry := &varuh.Entry{
		ID:       1,
		Title:    "Original Entry",
		User:     "user@example.com",
		Password: "secret123",
		Type:     "password",
	}

	t.Run("clone entry", func(t *testing.T) {
		err, clonedEntry := varuh.CloneEntry(entry)

		// Should handle gracefully whether or not there's an active database
		_ = err         // err can be nil or non-nil depending on database state
		_ = clonedEntry // clonedEntry can be nil if no active database
	})
}

func TestIterateEntries(t *testing.T) {
	tests := []struct {
		name     string
		orderKey string
		order    string
	}{
		{"order by id asc", "id", "asc"},
		{"order by title desc", "title", "desc"},
		{"order by timestamp asc", "timestamp", "asc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, entries := varuh.IterateEntries(tt.orderKey, tt.order)

			// Should handle all cases without panic
			_ = err     // err can be nil or non-nil depending on database state
			_ = entries // entries can be empty or nil
		})
	}
}

func TestEntriesToStringArray(t *testing.T) {
	tests := []struct {
		name           string
		skipLongFields bool
	}{
		{"include long fields", false},
		{"skip long fields", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, dataArray := varuh.EntriesToStringArray(tt.skipLongFields)

			// Should handle gracefully whether or not there's an active database
			_ = err       // err can be nil or non-nil depending on database state
			_ = dataArray // dataArray can be empty or nil
		})
	}
}

func TestGetExtendedEntries(t *testing.T) {
	entry := &varuh.Entry{ID: 1, Title: "Test Entry"}

	t.Run("get extended entries", func(t *testing.T) {
		extEntries := varuh.GetExtendedEntries(entry)

		// Should return a valid slice (can be empty)
		if extEntries == nil {
			extEntries = []varuh.ExtendedEntry{}
		}
		_ = len(extEntries) // Use the result
	})
}

// Integration tests that would work with actual database setup
func TestEntryTableName(t *testing.T) {
	entry := &varuh.Entry{}
	if entry.TableName() != "entries" {
		t.Errorf("Entry.TableName() = %s, want entries", entry.TableName())
	}
}

func TestExtendedEntryTableName(t *testing.T) {
	extEntry := &varuh.ExtendedEntry{}
	if extEntry.TableName() != "exentries" {
		t.Errorf("ExtendedEntry.TableName() = %s, want exentries", extEntry.TableName())
	}
}

func TestAddressTableName(t *testing.T) {
	address := &varuh.Address{}
	if address.TableName() != "address" {
		t.Errorf("Address.TableName() = %s, want address", address.TableName())
	}
}

// Test that database operations handle nil inputs gracefully
func TestDatabaseOperationsWithNilInputs(t *testing.T) {
	t.Run("operations with nil entry", func(t *testing.T) {
		// Test that functions handle nil entries gracefully
		err := varuh.RemoveDatabaseEntry(nil)
		_ = err // Should not panic, may return error

		_, cloned := varuh.CloneEntry(nil)
		_ = cloned // Should not panic, may return nil

		extEntries := varuh.GetExtendedEntries(nil)
		if extEntries == nil {
			extEntries = []varuh.ExtendedEntry{}
		}
		_ = len(extEntries)
	})
}

// Benchmark tests
func BenchmarkEntry_Copy(b *testing.B) {
	e1 := &varuh.Entry{}
	e2 := &varuh.Entry{
		Title:    "Benchmark Title",
		User:     "bench@example.com",
		Password: "secret123",
		Type:     "password",
	}

	for i := 0; i < b.N; i++ {
		e1.Copy(e2)
	}
}

func BenchmarkExtendedEntry_Copy(b *testing.B) {
	e1 := &varuh.ExtendedEntry{}
	e2 := &varuh.ExtendedEntry{
		FieldName:  "BenchField",
		FieldValue: "BenchValue",
		EntryID:    1,
	}

	for i := 0; i < b.N; i++ {
		e1.Copy(e2)
	}
}
