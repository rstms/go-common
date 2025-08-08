package common

import (
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

func ViperKey(name string) string {
	if programName == nil {
		panic("go-common: function called before Init()")
	}
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

func ViperGetStringSlice(key string) []string {
	return viper.GetStringSlice(ViperKey(key))
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
	configUsed := viper.ConfigFileUsed()
	if configUsed != "" {
		programConfigFile = &configUsed
		if ViperGetBool("verbose") {
			log.Println("Using config file:", viper.ConfigFileUsed())
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
			if !Confirm(fmt.Sprintf("Create directory '%s'?", dir)) {
				return
			}
			err := os.Mkdir(dir, 0700)
			cobra.CheckErr(err)
		}
		file = filepath.Join(dir, "config.yaml")
	}
	if IsFile(file) {
		if !Confirm(fmt.Sprintf("Overwrite config file '%s'?", file)) {
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
