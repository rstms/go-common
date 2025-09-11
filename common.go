package common

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const Version = "0.2.37"

var LogFile *os.File

var ConfirmAcceptMessage = "Proceeding"
var ConfirmRejectMessage = "Cowardly refused"

// client program name and version used for viper prefix and logfile header
// these are pointers so we'll panic if Init has not been called
var programName *string
var programVersion *string
var configFilename string

// call Init if not using cobra
func Init(name, version, configFile string) {
	setName(name, version)
	configFilename = configFile
	initConfig()
}

// call Shutdown at exit if not using cobra
func Shutdown() {
	shutdown()
}

// internal shutdown
func shutdown() {
	closeLog()
}

// called by Init or CobraInitRoot
func setName(name, version string) {
	programName = &name
	programVersion = &version
}

func checkInit() {
	if programName == nil || programVersion == nil {
		panic("go-common: function called before Init or CobraInit")
	}
}

func ProgramName() string {
	checkInit()
	return *programName
}

func ProgramVersion() string {
	checkInit()
	return *programVersion
}

func ConfigDir() string {
	checkInit()
	dir, _ := filepath.Split(configFilename)
	return strings.TrimRight(dir, string(filepath.Separator))
}

func CheckErr(err error) {
	if err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func openLog() {
	filename := ViperGetString("logfile")
	LogFile = nil
	if filename == "stdout" || filename == "-" {
		log.SetOutput(os.Stdout)
	} else if filename == "stderr" || filename == "" {
		log.SetOutput(os.Stderr)
	} else {
		fp, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
		if err != nil {
			log.Fatalf("failed opening log file: %v", err)
		}
		LogFile = fp
		log.SetOutput(LogFile)
		log.SetPrefix(fmt.Sprintf("[%d] ", os.Getpid()))
		log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
		log.Printf("%s v%s startup\n", ProgramName(), ProgramVersion())
	}
	if ViperGetBool("debug") {
		log.SetFlags(log.Flags() | log.Llongfile)
	}
}

func closeLog() {
	if LogFile != nil {
		log.Println("shutdown")
		err := LogFile.Close()
		if err != nil {
			log.Fatalf("failed closing log file: %v", err)
		}
		LogFile = nil
	}
}

func FormatJSON(v any) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatalf("failed formatting JSON: %v", err)
	}
	return string(data)
}
