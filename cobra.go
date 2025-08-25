package common

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var rootCmd *cobra.Command
var viperPrefix string
var configFilename string
var logFile *os.File

func checkRootCmd(name string) {
	if rootCmd == nil {
		cobra.CheckErr(fmt.Errorf("%s called before CobraInit(rootCmd)", name))
	}
}

func checkCobraCmd(name string, cobraCmd interface{}) *cobra.Command {
	cmd, ok := cobraCmd.(*cobra.Command)
	if !ok {
		cobra.CheckErr(fmt.Errorf("%s: cobraCmd not *cobra.Command: %v", name, cobraCmd))
	}
	return cmd
}

func OptionKey(cobraCmd interface{}, key string) string {
	checkRootCmd("OptionKey")
	cmd := checkCobraCmd("OptionKey", cobraCmd)
	prefix := rootCmd.Name() + "."
	if cmd == rootCmd {
		prefix += cmd.Name() + "."
	}
	return strings.ReplaceAll(prefix+key, "-", "_")
}

func OptionSwitch(cobraCmd interface{}, name, flag, description string) {

	checkRootCmd("OptionSwitch")
	cmd := checkCobraCmd("OptionSwitch", cobraCmd)

	if cmd == rootCmd {
		if flag == "" {
			rootCmd.PersistentFlags().Bool(name, false, description)
		} else {
			rootCmd.PersistentFlags().BoolP(name, flag, false, description)
		}
		viper.BindPFlag(OptionKey(cmd, name), rootCmd.PersistentFlags().Lookup(name))
	} else {
		if flag == "" {
			cmd.Flags().Bool(name, false, description)
		} else {
			cmd.Flags().BoolP(name, flag, false, description)
		}
		viper.BindPFlag(OptionKey(cmd, name), cmd.Flags().Lookup(name))
	}
}

func OptionString(cobraCmd interface{}, name, flag, defaultValue, description string) {

	checkRootCmd("OptionString")
	cmd := checkCobraCmd("OptionString", cobraCmd)

	if cmd == rootCmd {
		if flag == "" {
			rootCmd.PersistentFlags().String(name, defaultValue, description)
		} else {
			rootCmd.PersistentFlags().StringP(name, flag, defaultValue, description)
		}

		viper.BindPFlag(OptionKey(cmd, name), rootCmd.PersistentFlags().Lookup(name))
	} else {
		if flag == "" {
			cmd.PersistentFlags().String(name, defaultValue, description)
		} else {
			cmd.PersistentFlags().StringP(name, flag, defaultValue, description)
		}
		viper.BindPFlag(OptionKey(cmd, name), cmd.PersistentFlags().Lookup(name))
	}
}

func openLog() {
	filename := viper.GetString("logfile")
	logFile = nil
	if filename == "stdout" || filename == "-" {
		log.SetOutput(os.Stdout)
	} else if filename == "stderr" || filename == "" {
		log.SetOutput(os.Stderr)
	} else {
		fp, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
		if err != nil {
			log.Fatalf("failed opening log file: %v", err)
		}
		logFile = fp
		log.SetOutput(logFile)
		log.SetPrefix(fmt.Sprintf("[%d] ", os.Getpid()))
		log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
		log.Printf("%s v%s startup\n", rootCmd.Name(), rootCmd.Version)
		cobra.OnFinalize(closeLog)
	}
	if viper.GetBool("debug") {
		log.SetFlags(log.Flags() | log.Lshortfile)
	}
}

func closeLog() {
	if logFile != nil {
		log.Println("shutdown")
		err := logFile.Close()
		cobra.CheckErr(err)
		logFile = nil
	}
}

func initConfig() {
	checkRootCmd("initConfig")
	name := strings.ToLower(rootCmd.Name())
	viper.SetEnvPrefix(name)
	viper.AutomaticEnv()
	if configFilename != "" {
		viper.SetConfigFile(configFilename)
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
	openLog()
	if viper.ConfigFileUsed() != "" && viper.GetBool("verbose") {
		log.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func CobraInit(cobraRootCmd interface{}) {
	var ok bool
	rootCmd, ok = cobraRootCmd.(*cobra.Command)
	if !ok {
		cobra.CheckErr(fmt.Errorf("CobraInit: cobraRootCmd not *cobra.Command: %v", cobraRootCmd))
	}
	Init(rootCmd.Name(), rootCmd.Version)
	cacheDir, err := os.UserCacheDir()
	cobra.CheckErr(err)
	defaultCacheDir, err := TildePath(filepath.Join(cacheDir, rootCmd.Name()))
	cobra.CheckErr(err)

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&configFilename, "config-file", "", "config file")
	OptionString(rootCmd, "logfile", "", "", "log filename")
	OptionSwitch(rootCmd, "verbose", "v", "enable status output")
	OptionSwitch(rootCmd, "debug", "d", "enable diagnostic output")
	OptionString(rootCmd, "cache-dir", "", defaultCacheDir, "cache directory")
}
