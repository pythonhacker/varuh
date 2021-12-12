// varuh - command line encryption program written in Go.

package main

import (
	"fmt"
	"github.com/pythonhacker/argparse"
	"os"
)

const VERSION = 0.3
const APP = "varuh"

const AUTHOR_INFO = `
AUTHORS
    Copyright (C) 2021 Anand B Pillai <abpillai@gmail.com>
`

type actionFunc func(string) error
type actionFunc2 func(string) (error, string)
type voidFunc func() error
type voidFunc2 func() (error, string)

// Structure to keep the options data
type CmdOption struct {
	Short   string
	Long    string
	Help    string
	Path    string
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
func genPass() (error, string) {
	var err error
	var passwd string

	err, passwd = generateStrongPassword()

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
	}

	flagsActions2Map := map[string]voidFunc2{
		"genpass": genPass,
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

	// Flag 2 actions
	for key, mappedFunc := range flagsActions2Map {
		if *optMap[key].(*bool) {
			mappedFunc()
			flag = true
			break
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

	stringOptions := []CmdOption{
		{"I", "init", "Initialize a new database", "<path>", ""},
		{"d", "decrypt", "Decrypt password database", "<path>", ""},
		{"C", "clone", "Clone an entry with <id>", "<id>", ""},
		{"R", "remove", "Remove an entry with <id>", "<id>", ""},
		{"U", "use-db", "Set <path> as active database", "<path>", ""},
		{"f", "find", "Search entries with <term>", "<term>", ""},
		{"E", "edit", "Edit entry by <id>", "<id>", ""},
		{"l", "list-entry", "List entry by <id>", "<id>", ""},
		{"x", "export", "Export all entries to <filename>", "<filename>", ""},
	}

	for _, opt := range stringOptions {
		optMap[opt.Long] = parser.String(opt.Short, opt.Long, &argparse.Options{Help: opt.Help, Path: opt.Path})
	}

	boolOptions := []CmdOption{
		{"e", "encrypt", "Encrypt the current database", "", ""},
		{"A", "add", "Add a new entry", "", ""},
		{"p", "path", "Show current database path", "", ""},
		{"a", "list-all", "List all entries in current database", "", ""},
		{"g", "genpass", "Generate a strong password of length from 12 - 16", "", ""},
		{"s", "show", "Show passwords when listing entries", "", ""},
		{"c", "copy", "Copy password to clipboard", "", ""},
		{"v", "version", "Show version information and exit", "", ""},
		{"h", "help", "Print this help message and exit", "", ""},
	}

	for _, opt := range boolOptions {
		optMap[opt.Long] = parser.Flag(string(opt.Short), opt.Long, &argparse.Options{Help: opt.Help})
	}

	return optMap
}

// Main routine
func main() {
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "-h")
	}

	parser := argparse.NewParser("varuh",
		"Password manager for the command line for Unix like operating systems",
		AUTHOR_INFO,
	)

	optMap := initializeCmdLine(parser)

	err := parser.Parse(os.Args)

	if err != nil {
		fmt.Println(parser.Usage(err))
	}

	getOrCreateLocalConfig(APP)

	performAction(optMap)
}
