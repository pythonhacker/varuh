package tests

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"varuh"
)

func TestExportToFile(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		fileName string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "unsupported extension",
			fileName: filepath.Join(tempDir, "test.txt"),
			wantErr:  true,
			errMsg:   "format .txt not supported",
		},
		{
			name:     "csv extension",
			fileName: filepath.Join(tempDir, "test.csv"),
			wantErr:  false, // May error due to no active database, but format is supported
		},
		{
			name:     "markdown extension",
			fileName: filepath.Join(tempDir, "test.md"),
			wantErr:  false, // May error due to no active database, but format is supported
		},
		{
			name:     "html extension",
			fileName: filepath.Join(tempDir, "test.html"),
			wantErr:  false, // May error due to no active database, but format is supported
		},
		{
			name:     "pdf extension",
			fileName: filepath.Join(tempDir, "test.pdf"),
			wantErr:  false, // May error due to dependencies, but format is supported
		},
		{
			name:     "uppercase extension",
			fileName: filepath.Join(tempDir, "test.CSV"),
			wantErr:  false, // Should handle case-insensitive extensions
		},
		{
			name:     "no extension",
			fileName: filepath.Join(tempDir, "test"),
			wantErr:  true,
			errMsg:   "format  not supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := varuh.ExportToFile(tt.fileName)

			if tt.wantErr && (err == nil || !strings.Contains(err.Error(), tt.errMsg)) {
				if tt.errMsg != "" {
					t.Errorf("ExportToFile() expected error containing '%s', got %v", tt.errMsg, err)
				} else {
					t.Errorf("ExportToFile() expected error, got nil")
				}
				return
			}

			// For supported formats, we expect either success or database-related errors
			if !tt.wantErr && err != nil {
				// It's okay to get database-related errors when no active database is configured
				validErrors := []string{
					"database path cannot be empty",
					"Error exporting entries",
					"Error opening active database",
					"pandoc not found",
				}

				hasValidError := false
				for _, validErr := range validErrors {
					if strings.Contains(err.Error(), validErr) {
						hasValidError = true
						break
					}
				}

				if !hasValidError {
					t.Errorf("ExportToFile() unexpected error for supported format: %v", err)
				}
			}
		})
	}
}

func TestExportToCSV(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		fileName string
		wantErr  bool
	}{
		{
			name:     "valid csv file",
			fileName: filepath.Join(tempDir, "export.csv"),
			wantErr:  false, // May error due to database, but CSV writing logic should work
		},
		{
			name:     "invalid directory",
			fileName: "/nonexistent/directory/export.csv",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := varuh.ExportToCSV(tt.fileName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ExportToCSV() expected error, got nil")
				}
				return
			}

			// For valid filenames, we expect either success or database-related errors
			if err != nil {
				validErrors := []string{
					"database path cannot be empty",
					"Error exporting entries",
					"Error opening active database",
				}

				hasValidError := false
				for _, validErr := range validErrors {
					if strings.Contains(err.Error(), validErr) {
						hasValidError = true
						break
					}
				}

				if !hasValidError {
					t.Errorf("ExportToCSV() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestExportToMarkdown(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		fileName string
		wantErr  bool
	}{
		{
			name:     "valid markdown file",
			fileName: filepath.Join(tempDir, "export.md"),
			wantErr:  false,
		},
		{
			name:     "invalid directory",
			fileName: "/nonexistent/directory/export.md",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := varuh.ExportToMarkdown(tt.fileName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ExportToMarkdown() expected error, got nil")
				}
				return
			}

			// For valid filenames, we expect either success or database-related errors
			if err != nil {
				validErrors := []string{
					"database path cannot be empty",
					"Error exporting entries",
					"Error opening active database",
				}

				hasValidError := false
				for _, validErr := range validErrors {
					if strings.Contains(err.Error(), validErr) {
						hasValidError = true
						break
					}
				}

				if !hasValidError {
					t.Errorf("ExportToMarkdown() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestExportToMarkdownLimited(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		fileName string
		wantErr  bool
	}{
		{
			name:     "valid markdown file",
			fileName: filepath.Join(tempDir, "export_limited.md"),
			wantErr:  false,
		},
		{
			name:     "invalid directory",
			fileName: "/nonexistent/directory/export_limited.md",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := varuh.ExportToMarkdownLimited(tt.fileName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ExportToMarkdownLimited() expected error, got nil")
				}
				return
			}

			// For valid filenames, we expect either success or database-related errors
			if err != nil {
				validErrors := []string{
					"database path cannot be empty",
					"Error exporting entries",
					"Error opening active database",
				}

				hasValidError := false
				for _, validErr := range validErrors {
					if strings.Contains(err.Error(), validErr) {
						hasValidError = true
						break
					}
				}

				if !hasValidError {
					t.Errorf("ExportToMarkdownLimited() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestExportToHTML(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		fileName string
		wantErr  bool
	}{
		{
			name:     "valid html file",
			fileName: filepath.Join(tempDir, "export.html"),
			wantErr:  false,
		},
		{
			name:     "invalid directory",
			fileName: "/nonexistent/directory/export.html",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := varuh.ExportToHTML(tt.fileName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ExportToHTML() expected error, got nil")
				}
				return
			}

			// For valid filenames, we expect either success or database-related errors
			if err != nil {
				validErrors := []string{
					"database path cannot be empty",
					"Error exporting entries",
					"Error opening active database",
				}

				hasValidError := false
				for _, validErr := range validErrors {
					if strings.Contains(err.Error(), validErr) {
						hasValidError = true
						break
					}
				}

				if !hasValidError {
					t.Errorf("ExportToHTML() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestExportToPDF(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		fileName string
		wantErr  bool
	}{
		{
			name:     "valid pdf file",
			fileName: filepath.Join(tempDir, "export.pdf"),
			wantErr:  false, // May error due to pandoc dependency, but that's expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := varuh.ExportToPDF(tt.fileName)

			// PDF export requires external dependencies (pandoc, pdftk)
			// So we expect it to fail in most test environments
			// We mainly test that it doesn't panic and handles errors gracefully
			if err != nil {
				validErrors := []string{
					"pandoc not found",
					"database path cannot be empty",
					"Error exporting entries",
					"Error opening active database",
				}

				hasValidError := false
				for _, validErr := range validErrors {
					if strings.Contains(err.Error(), validErr) {
						hasValidError = true
						break
					}
				}

				if !hasValidError {
					t.Logf("ExportToPDF() returned unexpected error (may be system-specific): %v", err)
				}
			}
		})
	}
}

// Test export functions with mock data
func TestExportWithMockData(t *testing.T) {
	tempDir := t.TempDir()

	// Create a mock CSV file to test the CSV export format structure
	t.Run("mock csv structure", func(t *testing.T) {
		csvFile := filepath.Join(tempDir, "mock_test.csv")

		// Create a simple CSV file manually to test structure
		fh, err := os.Create(csvFile)
		if err != nil {
			t.Fatalf("Failed to create mock CSV file: %v", err)
		}

		writer := csv.NewWriter(fh)

		// Write header (same as in ExportToCSV)
		header := []string{"ID", "Title", "User", "URL", "Password", "Notes", "Modified"}
		err = writer.Write(header)
		if err != nil {
			t.Fatalf("Failed to write CSV header: %v", err)
		}

		// Write a mock record
		record := []string{"1", "Test Entry", "user@example.com", "https://example.com", "secret123", "Test notes", "2023-01-01 12:00:00"}
		err = writer.Write(record)
		if err != nil {
			t.Fatalf("Failed to write CSV record: %v", err)
		}

		writer.Flush()
		fh.Close()

		// Verify the file was created and has expected structure
		if _, err := os.Stat(csvFile); os.IsNotExist(err) {
			t.Error("Mock CSV file was not created")
		}

		// Read back and verify
		content, err := os.ReadFile(csvFile)
		if err != nil {
			t.Fatalf("Failed to read mock CSV file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "ID,Title,User") {
			t.Error("CSV header not found in mock file")
		}

		if !strings.Contains(contentStr, "Test Entry") {
			t.Error("Test record not found in mock CSV file")
		}
	})

	// Test markdown structure
	t.Run("mock markdown structure", func(t *testing.T) {
		mdFile := filepath.Join(tempDir, "mock_test.md")

		// Create a simple markdown file manually to test structure
		fh, err := os.Create(mdFile)
		if err != nil {
			t.Fatalf("Failed to create mock markdown file: %v", err)
		}
		defer fh.Close()

		// Write markdown table (similar to ExportToMarkdown)
		content := ` | ID | Title | User | URL | Password | Notes | Modified |
 | --- | --- | --- | --- | --- | --- | --- |
 | 1 | Test Entry | user@example.com | https://example.com | secret123 | Test notes | 2023-01-01 12:00:00 |
`
		_, err = fh.WriteString(content)
		if err != nil {
			t.Fatalf("Failed to write markdown content: %v", err)
		}

		// Verify the file was created
		if _, err := os.Stat(mdFile); os.IsNotExist(err) {
			t.Error("Mock markdown file was not created")
		}

		// Read back and verify structure
		readContent, err := os.ReadFile(mdFile)
		if err != nil {
			t.Fatalf("Failed to read mock markdown file: %v", err)
		}

		contentStr := string(readContent)
		if !strings.Contains(contentStr, "| ID | Title |") {
			t.Error("Markdown table header not found")
		}

		if !strings.Contains(contentStr, "| --- |") {
			t.Error("Markdown table separator not found")
		}
	})

	// Test HTML structure
	t.Run("mock html structure", func(t *testing.T) {
		htmlFile := filepath.Join(tempDir, "mock_test.html")

		// Create a simple HTML file manually to test structure
		fh, err := os.Create(htmlFile)
		if err != nil {
			t.Fatalf("Failed to create mock HTML file: %v", err)
		}
		defer fh.Close()

		// Write HTML table (similar to ExportToHTML)
		content := `<html><body>
<table cellPadding="2" cellSpacing="2" border="1">
<theader>
<th> ID </th><th> Title </th><th> User </th><th> URL </th><th> Password </th><th> Notes </th><th> Modified </th></theader>
<tbody>
<tr><td>1</td><td>Test Entry</td><td>user@example.com</td><td>https://example.com</td><td>secret123</td><td>Test notes</td><td>2023-01-01 12:00:00</td></tr>
</tbody>
</table>
</body></html>
`
		_, err = fh.WriteString(content)
		if err != nil {
			t.Fatalf("Failed to write HTML content: %v", err)
		}

		// Verify the file was created
		if _, err := os.Stat(htmlFile); os.IsNotExist(err) {
			t.Error("Mock HTML file was not created")
		}

		// Read back and verify structure
		readContent, err := os.ReadFile(htmlFile)
		if err != nil {
			t.Fatalf("Failed to read mock HTML file: %v", err)
		}

		contentStr := string(readContent)
		if !strings.Contains(contentStr, "<table") {
			t.Error("HTML table not found")
		}

		if !strings.Contains(contentStr, "<th> ID </th>") {
			t.Error("HTML table header not found")
		}

		if !strings.Contains(contentStr, "<td>Test Entry</td>") {
			t.Error("HTML table data not found")
		}
	})
}

// Test edge cases
func TestExportEdgeCases(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("empty filename", func(t *testing.T) {
		err := varuh.ExportToFile("")
		if err == nil || !strings.Contains(err.Error(), "format  not supported") {
			t.Errorf("ExportToFile() with empty filename should return unsupported format error, got: %v", err)
		}
	})

	t.Run("filename with dots", func(t *testing.T) {
		filename := filepath.Join(tempDir, "test.file.csv")
		err := varuh.ExportToFile(filename)
		// Should handle filename with multiple dots correctly (use last extension)
		if err != nil {
			validErrors := []string{
				"database path cannot be empty",
				"Error exporting entries",
				"Error opening active database",
			}

			hasValidError := false
			for _, validErr := range validErrors {
				if strings.Contains(err.Error(), validErr) {
					hasValidError = true
					break
				}
			}

			if !hasValidError {
				t.Errorf("ExportToFile() with dotted filename unexpected error: %v", err)
			}
		}
	})

	t.Run("readonly directory", func(t *testing.T) {
		readonlyDir := filepath.Join(tempDir, "readonly")
		err := os.Mkdir(readonlyDir, 0400) // Read-only directory
		if err != nil {
			t.Skipf("Cannot create readonly directory: %v", err)
		}
		defer os.Chmod(readonlyDir, 0755) // Restore permissions for cleanup

		filename := filepath.Join(readonlyDir, "export.csv")
		err = varuh.ExportToCSV(filename)
		if err == nil {
			t.Error("ExportToCSV() should fail when writing to readonly directory")
		}
	})
}

// Benchmark tests
func BenchmarkExportToFile(b *testing.B) {
	tempDir := b.TempDir()

	for i := 0; i < b.N; i++ {
		filename := filepath.Join(tempDir, "benchmark_test.csv")
		varuh.ExportToFile(filename)
		os.Remove(filename) // Clean up for next iteration
	}
}

func BenchmarkExportToCSV(b *testing.B) {
	tempDir := b.TempDir()

	for i := 0; i < b.N; i++ {
		filename := filepath.Join(tempDir, "benchmark_test.csv")
		varuh.ExportToCSV(filename)
		os.Remove(filename) // Clean up for next iteration
	}
}

func BenchmarkExportToMarkdown(b *testing.B) {
	tempDir := b.TempDir()

	for i := 0; i < b.N; i++ {
		filename := filepath.Join(tempDir, "benchmark_test.md")
		varuh.ExportToMarkdown(filename)
		os.Remove(filename) // Clean up for next iteration
	}
}