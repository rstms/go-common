package common

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const Version = "0.1.3"

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

func ViperKey(name string) string {
	var prefix string
	if *programName != "" {
		prefix = *programName + "."
	}
	return strings.ToLower(strings.ReplaceAll(prefix+name, "-", "_"))
}

func ViperGetBool(key string) bool {
	return viper.GetBool(ViperKey(key))
}

func ViperGetString(key string) string {
	return Expand(viper.GetString(ViperKey(key)))
}

func ViperGetInt(key string) int {
	return viper.GetInt(ViperKey(key))
}

func ViperGetInt64(key string) int64 {
	return viper.GetInt64(ViperKey(key))
}

func ViperSet(key string, value any) {
	viper.Set(ViperKey(key), value)
}

func ViperSetDefault(key string, value any) {
	viper.SetDefault(ViperKey(key), value)
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

func Expand(pathname string) string {
	if len(pathname) > 1 && pathname[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("failed getting user home dir: %v", err)
		}
		pathname = filepath.Join(home, pathname[1:])
	}
	pathname = os.ExpandEnv(pathname)
	return pathname
}

func InitConfig(configFile string) {
	programConfigFile = &configFile
	name := strings.ToLower(strings.ReplaceAll(*programName, "-", "_"))
	viper.SetEnvPrefix(name)
	viper.AutomaticEnv()
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		userConfig, err := os.UserConfigDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(filepath.Join(home, "."+name))
		viper.AddConfigPath(filepath.Join(userConfig, name))
		viper.AddConfigPath(filepath.Join("/etc", name))
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}
	err := viper.ReadInConfig()
	if err != nil {
		_, ok := err.(viper.ConfigFileNotFoundError)
		if !ok {
			cobra.CheckErr(err)
		}
	}
	OpenLog()
	if viper.ConfigFileUsed() != "" && ViperGetBool("verbose") {
		log.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func Confirm(prompt string) bool {
	if ViperGetBool("force") {
		return true
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s [y/N]: ", prompt)
		response, err := reader.ReadString('\n')
		cobra.CheckErr(err)
		response = strings.ToLower(strings.TrimSpace(response))
		if response == "y" || response == "yes" {
			msg := ViperGetString("messages.confirm_accept")
			if msg == "" {
				msg = ConfirmAcceptMessage
			}
			if msg != "" {
				fmt.Println(msg)
			}
			return true
		} else if response == "n" || response == "no" || response == "" {
			msg := ViperGetString("messages.confirm_reject")
			if msg == "" {
				msg = ConfirmRejectMessage
			}
			if msg != "" {
				fmt.Println(msg)
			}
			return false
		}
	}
}

func InitConfigFile() {
	file := *programConfigFile
	if file == "" {
		userConfig, err := os.UserConfigDir()
		cobra.CheckErr(err)
		dir := filepath.Join(userConfig, strings.ToLower(strings.ReplaceAll(*programName, "-", "_")))
		if !IsDir(dir) {
			if !Confirm(fmt.Sprintf("Create directory '%s'", dir)) {
				return
			}
			err := os.Mkdir(dir, 0700)
			cobra.CheckErr(err)
		}
		file = filepath.Join(dir, "config.yaml")
	}
	if IsFile(file) {
		if !Confirm(fmt.Sprintf("Overwrite config file '%s'", file)) {
			return
		}
	}
	err := viper.WriteConfigAs(file)
	cobra.CheckErr(err)
	fmt.Printf("Default configuration written to %s\n", file)
}

func EditConfigFile() {
	var editCommand string
	if runtime.GOOS == "windows" {
		editCommand = "notepad"
	} else {
		editCommand = os.Getenv("VISUAL")
		if editCommand == "" {
			editCommand = os.Getenv("EDITOR")
			if editCommand == "" {
				editCommand = "vi"
			}
		}
	}
	editor := exec.Command(editCommand, viper.ConfigFileUsed())
	editor.Stdin = os.Stdin
	editor.Stdout = os.Stdout
	editor.Stderr = os.Stderr
	err := editor.Run()
	cobra.CheckErr(err)
}
