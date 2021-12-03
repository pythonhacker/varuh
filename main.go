// varuh - command line encryption program written in Go.

package main

import (
	"fmt"
	"strconv"
	getopt "github.com/pborman/getopt/v2"
	"os"
)

const VERSION = 0.2
const APP = "varuh"
const AUTHOR_EMAIL = "Anand B Pillai <abpillai@gmail.com>"

type actionFunc func(string) error
type actionFunc2 func(string) (error, string)
type voidFunc func() error

// Print the program's usage string and exit
func printUsage() error {
	getopt.Usage()
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

// Perform an action by using the command line options map
func performAction(optMap map[string]interface{}, optionMap map[string]interface{}) {

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
		option := optionMap[key].(Option)

		if *optMap[key].(*string) != option.Path {

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
		option := optionMap[key].(Option)

		if *optMap[key].(*string) != option.Path {

			var val = *(optMap[key].(*string))
			mappedFunc(val)
			break
		}
	}

}

// Main routine
func main() {
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "-h")
	}

	optMap, optionMap := initializeCommandLine()
	getopt.SetUsage(func() {
		usageString(optionMap)
	})

	getopt.Parse()
	getOrCreateLocalConfig(APP)

	performAction(optMap, optionMap)
}
