package common

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	prefix := ProgramName() + "."
	if cmd != rootCmd {
		prefix += cmd.Name() + "."
	}
	return strings.ToLower(strings.ReplaceAll(prefix+key, "-", "_"))
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

func OptionInt(cobraCmd CobraCommand, name, flag string, defaultValue int, description string) {

	checkRootCmd("OptionInt")
	cmd := toCobraCmd("OptionInt", "cobraCmd", cobraCmd)

	if cmd == rootCmd {
		if flag == "" {
			rootCmd.PersistentFlags().Int(name, defaultValue, description)
		} else {
			rootCmd.PersistentFlags().IntP(name, flag, defaultValue, description)
		}

		viper.BindPFlag(OptionKey(cmd, name), rootCmd.PersistentFlags().Lookup(name))
	} else {
		if flag == "" {
			cmd.PersistentFlags().Int(name, defaultValue, description)
		} else {
			cmd.PersistentFlags().IntP(name, flag, defaultValue, description)
		}
		viper.BindPFlag(OptionKey(cmd, name), cmd.PersistentFlags().Lookup(name))
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
	switch rootCmd {
	case nil:
		rootCmd = root
	case root:
		// rootCmd has already been set by a call to CobraAddCommand
		return
	default:
		// rootCmd must match if non-nil
		cobra.CheckErr(fmt.Errorf("CobraInit: rootCmd mismatch; got %v, expected %v", root, rootCmd))
	}

	// rootCmd initialization
	setName(rootCmd.Name(), rootCmd.Version)
	cacheDir, err := os.UserCacheDir()
	cobra.CheckErr(err)

	defaultCacheDir, err := TildePath(filepath.Join(cacheDir, rootCmd.Name()))
	cobra.CheckErr(err)

	cobra.OnInitialize(initConfig)
	cobra.OnFinalize(shutdown)

	rootCmd.PersistentFlags().StringVar(&configFilename, "config-file", "", "config file")
	OptionString(rootCmd, "logfile", "l", "stderr", "log filename")
	OptionSwitch(rootCmd, "verbose", "v", "enable status output")
	OptionSwitch(rootCmd, "debug", "d", "enable diagnostic output")
	OptionString(rootCmd, "cache-dir", "", defaultCacheDir, "cache directory")

	CobraAddCommand(rootCmd, rootCmd, configCmd)
	OptionSwitch(configCmd, "no-header", "", "suppress config header comments")
	CobraAddCommand(rootCmd, configCmd, configCatCmd)
	CobraAddCommand(rootCmd, configCmd, configEditCmd)
	CobraAddCommand(rootCmd, configCmd, configFileCmd)
	CobraAddCommand(rootCmd, configCmd, configInitCmd)
}
