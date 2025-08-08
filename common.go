package common

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
)

const Version = "0.1.18"

var LogFile *os.File

var ConfirmAcceptMessage = "Proceeding"
var ConfirmRejectMessage = "Cowardly refused"

// client program name and version used for viper prefix and logfile header
// these are pointers so we'll panic if Init has not been called
var programName *string
var programVersion *string
var programConfigFile *string

// must be called before any other functions
func Init(name, version string) {
	programName = &name
	programVersion = &version
}

func ProgramName() string {
	return *programName
}

func ProgramVersion() string {
	return *programVersion
}

func OpenLog() {
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
		log.Printf("%s v%s startup\n", *programName, *programVersion)
		cobra.OnFinalize(CloseLog)
	}
	if ViperGetBool("debug") {
		log.SetFlags(log.Flags() | log.Lshortfile)
	}
}

func CloseLog() {
	if LogFile != nil {
		log.Println("shutdown")
		err := LogFile.Close()
		cobra.CheckErr(err)
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

func IsDir(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

func IsFile(pathname string) bool {
	_, err := os.Stat(pathname)
	return !os.IsNotExist(err)
}
