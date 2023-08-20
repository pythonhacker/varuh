package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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
