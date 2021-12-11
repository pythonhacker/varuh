// varuh - command line encryption program written in Go.

package main

import (
	"fmt"
	"strconv"
	//	getopt "github.com/pborman/getopt/v2"
	"github.com/akamensky/argparse"	
	"os"
)

const VERSION = 0.2
const APP = "varuh"
const AUTHOR_EMAIL = "Anand B Pillai <abpillai@gmail.com>"

type actionFunc func(string) error
type actionFunc2 func(string) (error, string)
type voidFunc func() error

// Structure to keep the options data
type CmdOption struct {
	Short string
	Long  string
	Help  string
	Default string
}

// Print the program's usage string and exit
func printUsage() error {
	//	getopt.Usage()
	os.Exit(0)

	return nil
}

// Print the program's version info and exit
func printVersionInfo() error {
	fmt.Printf("%s version %.2f\n", APP, VERSION)
	os.Exit(0)

	return nil
}

// Command-line wrapper to generateRandomPassword
func generatePassword(length string) (error, string) {
	var iLength int
	var err error
	var passwd string
	
	iLength, _ = strconv.Atoi(length)
	err, passwd = generateRandomPassword(iLength)

	if err != nil {
		fmt.Printf("Error generating password - \"%s\"\n", err.Error())
		return err, ""
	}

	fmt.Println(passwd)

	if settingsRider.CopyPassword {
		copyPasswordToClipboard(passwd)
		fmt.Println("Password copied to clipboard")
	}
	
	return nil, passwd
}

// // Perform an action by using the command line options map
func performAction(optMap map[string]interface{}) {

	var flag bool

	boolActionsMap := map[string]voidFunc{
		"add":      WrapperMaxKryptVoidFunc(addNewEntry),
		"version":  printVersionInfo,
		"help":     printUsage,
		"path":     showActiveDatabasePath,
		"list-all": WrapperMaxKryptVoidFunc(listAllEntries),
		"encrypt":  encryptActiveDatabase,
	}

	stringActionsMap := map[string]actionFunc{
		"edit":       WrapperMaxKryptStringFunc(editCurrentEntry),
		"init":       initNewDatabase,
		"list-entry": WrapperMaxKryptStringFunc(listCurrentEntry),
		"find":       WrapperMaxKryptStringFunc(findCurrentEntry),
		"remove":     WrapperMaxKryptStringFunc(removeCurrentEntry),
		"clone":      WrapperMaxKryptStringFunc(copyCurrentEntry),
		"use-db":     setActiveDatabasePath,
		"export":     exportToFile,
	}

	stringActions2Map := map[string]actionFunc2{
		"decrypt": decryptDatabase,
		"genpass": generatePassword,		
	}

	flagsActionsMap := map[string]voidFunc{
		"show": setShowPasswords,
		"copy": setCopyPasswordToClipboard,
	}

	// Flag actions - always done
	for key, mappedFunc := range flagsActionsMap {
		if *optMap[key].(*bool) {
			mappedFunc()
		}
	}

	
	// One of bool or string actions
	for key, mappedFunc := range boolActionsMap {
		if *optMap[key].(*bool) {
			mappedFunc()
			flag = true
			break
		}
	}

	if flag {
		return
	}

	for key, mappedFunc := range stringActionsMap {
		if *optMap[key].(*string) != "" {

			var val = *(optMap[key].(*string))
			mappedFunc(val)
			flag = true
			break
		}
	}

	if flag {
		return
	}

	for key, mappedFunc := range stringActions2Map {
		if *optMap[key].(*string) != "" {
			var val = *(optMap[key].(*string))
			mappedFunc(val)
			break
		}
	}

}

func initializeCmdLine(parser *argparse.Parser) map[string]interface{} {
	var optMap map[string]interface{}

	optMap = make(map[string]interface{})

	boolOptions := []CmdOption{
		{"e", "encrypt", "Encrypt the current database", ""},
		{"A", "add", "Add a new entry", ""},
		{"p", "path", "Show current database path", ""},
		{"a", "list-all", "List all entries in current database", ""},
		{"s", "show", "Show passwords when listing entries", ""},
		{"c", "copy", "Copy password to clipboard", ""},
		{"v", "version", "Show version information and exit", ""},
		{"h", "help", "Print this help message and exit", ""},
	}

	for _, opt := range boolOptions {
		optMap[opt.Long] = parser.Flag(string(opt.Short), opt.Long, &argparse.Options{Help: opt.Help})
	}	

	stringOptions := []CmdOption{
		{"I", "init", "Initialize a new database", ""},
		{"d", "decrypt", "Decrypt password database", ""},
		{"C", "clone", "Clone an entry", ""},
		{"R", "remove", "Remove an entry", ""},
		{"U", "use-db", "Set as active database", ""},
		{"f", "find", "Search entries", ""},
		{"E", "edit", "Edit entry by id", ""},
		{"l", "list-entry", "List entry by id", ""},
		{"x", "export", "Export all entries to <filename>", ""},
		{"g", "genpass", "Generate password of given length", "12"},				
	}

	for _, opt := range stringOptions {
		optMap[opt.Long] = parser.String(opt.Short, opt.Long, &argparse.Options{Help: opt.Help, Default: opt.Default})
	}
	
	return optMap
}

// Main routine
func main() {
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "-h")
	}

	parser := argparse.NewParser("varuh", "Password manager for the command line for Unix like operating systems")
		
	//	optMap, optionMap := initializeCommandLine(parser)

	//	versionFlag := parser.Flag("v", "version", &argparse.Options{Help: "Show version information and exit"})
	optMap := initializeCmdLine(parser)
	
	err := parser.Parse(os.Args)

	if err != nil {
		fmt.Println(parser.Usage(err))
	}
	
	getOrCreateLocalConfig(APP)

	performAction(optMap)
}
