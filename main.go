// xkuz - command line encryption program written in Go.

package main

import (
	"fmt"
	getopt "github.com/pborman/getopt/v2"
	"os"
)

const VERSION = 0.1
const APP = "varuh"
const AUTHOR_EMAIL = "Anand B Pillai <anandpillai@alumni.iitm.ac.in>"

type actionFunc func(string) error
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

// Perform an action by using the command line options map
func performAction(optMap map[string]interface{}, optionMap map[string]interface{}) {

	var flag bool

	boolActionsMap := map[string]voidFunc{
		"add":      addNewEntry,
		"version":  printVersionInfo,
		"help":     printUsage,
		"path":     showActiveDatabasePath,
		"list-all": listAllEntries,
		"encrypt":  encryptActiveDatabase,
	}

	stringActionsMap := map[string]actionFunc{
		"edit":       editCurrentEntry,
		"init":       initNewDatabase,
		"list-entry": listCurrentEntry,
		"find":       findCurrentEntry,
		"remove":     removeCurrentEntry,
		"copy":       copyCurrentEntry,
		"use-db":     setActiveDatabasePath,
		"decrypt":    decryptDatabase,
	}

	flagsActionsMap := map[string]voidFunc{
		"show": setShowPasswords,
	}

	// Flag actions - always done
	for key, mappedFunc := range flagsActionsMap {
		if *optMap[key].(*bool) {
			mappedFunc()
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
