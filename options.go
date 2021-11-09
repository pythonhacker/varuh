// Managing command line options
package main

import (
	"fmt"
	"strings"

	getopt "github.com/pborman/getopt/v2"
)

// Structure to keep the options data
type Option struct {
	Short rune
	Long  string
	Path  string
	Help  string
	Type  uint8
}

// Usage string template
const HELP_STRING = `
SYNOPSIS

    %s [options] [flags]

OPTIONS

    EDIT/CREATE ACTIONS:

%s

    FIND/LIST ACTIONS:

%s

    HELP ACTIONS:

%s

    FLAGS:

%s


AUTHORS
    Copyright (C) 2021 %s
`

// Generate the usage string for the program
func usageString(optMap map[string]interface{}) {

	var editActions []string
	var findActions []string
	var helpActions []string
	var flagActions []string

	var maxLen1 int
	var maxLen2 int

	var usageTemplate = "%8s --%s %s %s"

	// Find max string length
	for _, value := range optMap {
		option := value.(Option)

		if len(option.Long) > maxLen1 {
			maxLen1 = len(option.Long)
		}
		if len(option.Path) > maxLen2 {
			maxLen2 = len(option.Path)
		}
	}

	for _, value := range optMap {
		option := value.(Option)

		delta := maxLen1 + 5 - len(option.Long)
		for i := 0; i < delta; i++ {
			option.Long += " "
		}

		if len(option.Path) < maxLen2 {
			delta := maxLen2 - len(option.Path)
			for i := 0; i < delta; i++ {
				option.Path += " "
			}
		}

		switch option.Type {
		case 0:
			editActions = append(editActions, fmt.Sprintf(usageTemplate, "-"+string(option.Short), option.Long, option.Path, option.Help))
		case 1:
			findActions = append(findActions, fmt.Sprintf(usageTemplate, "-"+string(option.Short), option.Long, option.Path, option.Help))
		case 2:
			helpActions = append(helpActions, fmt.Sprintf(usageTemplate, "-"+string(option.Short), option.Long, option.Path, option.Help))
		case 3:
			flagActions = append(flagActions, fmt.Sprintf(usageTemplate, "-"+string(option.Short), option.Long, option.Path, option.Help))
		}
	}

	fmt.Println(fmt.Sprintf(HELP_STRING, APP, strings.Join(editActions, "\n"),
		strings.Join(findActions, "\n"), strings.Join(helpActions, "\n"),
		strings.Join(flagActions, "\n"), AUTHOR_EMAIL))

}

// Set up command line options - returns two maps
func initializeCommandLine() (map[string]interface{}, map[string]interface{}) {
	var optMap map[string]interface{}
	var optionMap map[string]interface{}

	optMap = make(map[string]interface{})
	optionMap = make(map[string]interface{})

	stringOptions := []Option{
		{'I', "init", "<path>", "Initialize a new database", 0},
		{'d', "decrypt", "<path>", "Decrypt password database", 0},
		{'C', "copy", "<id>", "Copy an entry", 0},
		{'R', "remove", "<id>", "Remove an entry", 0},
		{'U', "use-db", "<path>", "Set as active database", 0},
		{'f', "find", "<term>", "Search entries", 1},
		//		{'r',"regex","<term>","Search entries using regular expressions", 1},
		{'E', "edit", "<id>", "Edit entry by id", 0},
		{'l', "list-entry", "<id>", "List entry by id", 1},
	}

	for _, opt := range stringOptions {
		optMap[opt.Long] = getopt.StringLong(opt.Long, opt.Short, opt.Path, opt.Help)
		optionMap[opt.Long] = opt
	}

	boolOptions := []Option{
		{'e', "encrypt", "", "Encrypt the current database", 0},
		{'A', "add", "", "Add a new entry", 0},
		{'p', "path", "", "Show current database path", 1},
		{'a', "list-all", "", "List all entries in current database", 1},
		{'s', "show", "", "Show passwords when listing entries", 3},
		{'v', "version", "", "Show version information and exit", 2},
		{'h', "help", "", "Print this help message and exit", 2},
	}

	for _, opt := range boolOptions {
		optMap[opt.Long] = getopt.BoolLong(opt.Long, opt.Short, opt.Help)
		optionMap[opt.Long] = opt
	}

	return optMap, optionMap
}
