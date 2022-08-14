// varuh - command line encryption program written in Go.

package main

import (
    "fmt"
    "github.com/pythonhacker/argparse"
    "os"
    "strings"
)

const VERSION = 0.4
const APP = "varuh"

const AUTHOR_INFO = `
AUTHORS
    Copyright (C) 2022 Anand B Pillai <abpillai@gmail.com>
`

type actionFunc func(string) error
type actionFunc2 func(string) (error, string)
type voidFunc func() error
type voidFunc2 func() (error, string)
type settingFunc func(string) 

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
    //  getopt.Usage()
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
        "remove":     WrapperMaxKryptStringFunc(removeCurrentEntry),
        "clone":      WrapperMaxKryptStringFunc(copyCurrentEntry),
        "use-db":     setActiveDatabasePath,
        "export":     exportToFile,
        "migrate":    migrateDatabase,
    }

    stringListActionsMap := map[string]actionFunc{
        "find": WrapperMaxKryptStringFunc(findCurrentEntry),
    }

    stringActions2Map := map[string]actionFunc2{
        "decrypt": decryptDatabase,
    }

    flagsActions2Map := map[string]voidFunc2{
        "genpass": genPass,
    }

    flagsActionsMap := map[string]voidFunc{
        "show":       setShowPasswords,
        "copy":       setCopyPasswordToClipboard,
        "assume-yes": setAssumeYes,
    }

    flagsSettingsMap := map[string]settingFunc{
        "type": setType,
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

    // Settings
    for key, mappedFunc := range flagsSettingsMap {
        if *optMap[key].(*string) != ""{
            var val = *(optMap[key].(*string))
            mappedFunc(val)         
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

    for key, mappedFunc := range stringListActionsMap {
        if len(*optMap[key].(*[]string)) > 0 {

            var vals = *(optMap[key].(*[]string))
            // Convert to single string
            var singleVal = strings.Join(vals, " ")
            mappedFunc(singleVal)
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
        {"R", "remove", "Remove an entry with <id> or <id-range>", "<id>", ""},
        {"U", "use-db", "Set <path> as active database", "<path>", ""},
        {"E", "edit", "Edit entry by <id>", "<id>", ""},
        {"l", "list-entry", "List entry by <id>", "<id>", ""},
        {"x", "export", "Export all entries to <filename>", "<filename>", ""},
        {"m", "migrate", "Migrate a database to latest schema", "<path>", ""},
        {"t", "type", "Specify type when adding a new entry", "<type>", ""},
    }

    for _, opt := range stringOptions {
        optMap[opt.Long] = parser.String(opt.Short, opt.Long, &argparse.Options{Help: opt.Help, Path: opt.Path})
    }

    stringListOptions := []CmdOption{
        {"f", "find", "Search entries with terms", "<t1> <t2> ...", ""},
    }

    for _, opt := range stringListOptions {
        optMap[opt.Long] = parser.StringList(opt.Short, opt.Long, &argparse.Options{Help: opt.Help, Path: opt.Path})
    }

    boolOptions := []CmdOption{
        {"e", "encrypt", "Encrypt the current database", "", ""},
        {"A", "add", "Add a new entry", "", ""},
        {"p", "path", "Show current database path", "", ""},
        {"a", "list-all", "List all entries in current database", "", ""},
        {"g", "genpass", "Generate a strong password (length: 12 - 16)", "", ""},
        {"s", "show", "Show passwords when listing entries", "", ""},
        {"c", "copy", "Copy password to clipboard", "", ""},
        {"y", "assume-yes", "Assume yes to actions requiring confirmation", "", ""},
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
