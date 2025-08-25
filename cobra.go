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

type CobraCommand interface {
}

var rootCmd *cobra.Command
var viperPrefix string
var logFile *os.File

// fail with informative error instead of panic if CobraInit has not been called
func checkRootCmd(name string) {
	if rootCmd == nil {
		cobra.CheckErr(fmt.Errorf("%s called before CobraInit", name))
	}
}

// cast CobraCommand to *cobra.Command with informative error on failure
func toCobraCmd(funcName, argName string, arg CobraCommand) *cobra.Command {
	cmd, ok := arg.(*cobra.Command)
	if !ok {
		cobra.CheckErr(fmt.Errorf("%s: %s type mismatch; expected *cobra.Command, got %v", funcName, argName, arg))
	}
	return cmd
}

func OptionKey(cobraCmd CobraCommand, key string) string {
	checkRootCmd("OptionKey")
	cmd := toCobraCmd("OptionKey", "cobraCmd", cobraCmd)
	prefix := rootCmd.Name() + "."
	if cmd == rootCmd {
		prefix += cmd.Name() + "."
	}
	return strings.ReplaceAll(prefix+key, "-", "_")
}

func OptionSwitch(cobraCmd CobraCommand, name, flag, description string) {

	checkRootCmd("OptionSwitch")
	cmd := toCobraCmd("OptionSwitch", "cobraCmd", cobraCmd)

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

func OptionString(cobraCmd CobraCommand, name, flag, defaultValue, description string) {

	checkRootCmd("OptionString")
	cmd := toCobraCmd("OptionString", "cobraCmd", cobraCmd)

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

// call from non-root cobra command init
func CobraAddCommand(cobraRootCmd, parentCmd, cobraCmd CobraCommand) {
	root := toCobraCmd("CobraAddCommand", "cobraRootCmd", cobraRootCmd)
	switch rootCmd {
	case nil:
		// CobraInit has not been called yet, so do so now
		CobraInit(cobraRootCmd)
	default:
		// if rootCmd already set, the argument must match
		if root != rootCmd {
			cobra.CheckErr(fmt.Errorf("CobraAddCommand: cobraRootCmd mismatch; got %v, expected %v", root, rootCmd))
		}
	}
	parent := toCobraCmd("CobraAddCommand", "parentCmd", parentCmd)
	cmd := toCobraCmd("CobraAddCommand", "cobraCmd", cobraCmd)
	parent.AddCommand(cmd)
}

// call from root cobra command init
func CobraInit(cobraRootCmd CobraCommand) {

	root := toCobraCmd("CobraInit", "cobraRootCmd", cobraRootCmd)
	if rootCmd != nil {
		// rootCmd has already been set by a call to CobraAddCommand
		if root == rootCmd {
			// the argument matches, we're done here
			return
		}
		cobra.CheckErr(fmt.Errorf("CobraInit: rootCmd mismatch; got %v, expected %v", root, rootCmd))
	}
	// rootCmd initialization
	setName(rootCmd.Name(), rootCmd.Version)
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
