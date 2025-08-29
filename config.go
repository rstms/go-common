/*
Copyright © 2025 Matt Krueger <mkrueger@rstms.net>
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

 1. Redistributions of source code must retain the above copyright notice,
    this list of conditions and the following disclaimer.

 2. Redistributions in binary form must reproduce the above copyright notice,
    this list of conditions and the following disclaimer in the documentation
    and/or other materials provided with the distribution.

 3. Neither the name of the copyright holder nor the names of its contributors
    may be used to endorse or promote products derived from this software
    without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/
package common

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v3"
)

func ConfigString(header bool) string {
	checkRootCmd("ConfigString")
	var config string
	if header {
		config += configHeader()
	}
	config += configYAML()
	return config
}

func configHeader() string {
	name := rootCmd.Name()
	buf := fmt.Sprintf("# %s config\n", name)
	if ViperGetBool("verbose") {
		currentUser, err := user.Current()
		cobra.CheckErr(err)
		hostname, err := os.Hostname()
		cobra.CheckErr(err)

		buf += fmt.Sprintf("# generated: %s by %s@%s (%s_%s)\n",
			time.Now().Format(time.DateTime),
			currentUser.Username, hostname,
			runtime.GOOS, runtime.GOARCH,
		)
	}

	userConfig, err := os.UserConfigDir()
	cobra.CheckErr(err)

	userConfig, err = TildePath(userConfig)
	cobra.CheckErr(err)

	buf += fmt.Sprintf("# default_config_dir: %s\n", filepath.Join(userConfig, name))

	userCache, err := os.UserCacheDir()
	cobra.CheckErr(err)

	userCache, err = TildePath(userCache)
	cobra.CheckErr(err)

	buf += fmt.Sprintf("# default_cache_dir: %s\n", filepath.Join(userCache, name))
	return buf
}

func configYAML() string {
	configMap := viper.AllSettings()
	//keys := viper.AllKeys()
	fmt.Printf("before: %s\n", FormatJSON(configMap))
	fmt.Printf("delete: %s\n", ProgramName()+".config")
	delete(configMap, ProgramName()+".config")
	/*
		for _, key := range keys {
			fmt.Printf("key: %s\n", key)

			if strings.HasPrefix(key, ProgramName()+".config") {
				fmt.Printf("  deleting: %s\n", key)
				delete(configMap, key)
			}
		}
	*/
	fmt.Printf("after: %s\n", FormatJSON(configMap))
	var buf bytes.Buffer
	func() {
		encoder := yaml.NewEncoder(&buf)
		defer encoder.Close()
		encoder.SetIndent(2)
		err := encoder.Encode(&configMap)
		cobra.CheckErr(err)
	}()
	return buf.String()
}

func ConfigInit(allowClobber bool) string {
	checkRootCmd("ConfigInit")
	configFilename := viper.ConfigFileUsed()
	switch configFilename {
	case "":
		userConfigDir, err := os.UserConfigDir()
		cobra.CheckErr(err)
		configDir := filepath.Join(userConfigDir, ProgramName())
		if !IsDir(configDir) {
			err := os.Mkdir(configDir, 0700)
			cobra.CheckErr(err)
		}
		configFilename = filepath.Join(configDir, "config.yaml")

	default:
		if !allowClobber {
			cobra.CheckErr(fmt.Errorf("not overwriting current file: %s\n", configFilename))
		}
	}

	configFile, err := os.Create(configFilename)
	cobra.CheckErr(err)
	defer configFile.Close()
	fmt.Fprintf(configFile, "%s\n%s\n", configHeader(), configYAML())
	return configFilename
}

func ConfigEdit() {
	checkRootCmd("ConfigEdit")
	configFilename := viper.ConfigFileUsed()
	if configFilename == "" {
		configFilename = ConfigInit(false)
	}
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
	editor := exec.Command(editCommand, configFilename)
	editor.Stdin = os.Stdin
	editor.Stdout = os.Stdout
	editor.Stderr = os.Stderr
	err := editor.Run()
	cobra.CheckErr(err)
}

func initConfig() {
	name := strings.ToLower(ProgramName())
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
	if viper.ConfigFileUsed() != "" && ViperGetBool("verbose") {
		log.Println("Using config file:", viper.ConfigFileUsed())
	}
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "config subcommands",
	Long: `
subcommand for viewing or modifying the config file
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ConfigString(!ViperGetBool("config.no_header")))
	},
}

var configCatCmd = &cobra.Command{
	Use:   "cat",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ConfigString(true))
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "edit the config file",
	Long: `
edit the config file using the system editor.  If no config file exists, 
create one in the default location before editing.
`,
	Run: func(cmd *cobra.Command, args []string) {
		ConfigEdit()
	},
}

var configFileCmd = &cobra.Command{
	Use:   "file",
	Short: "output the config file",
	Long: `
write the pathname of the active config file to stdout
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(viper.ConfigFileUsed())
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "initialize config file",
	Long: `
write a YAML config file to the default location
`,
	Run: func(cmd *cobra.Command, args []string) {
		ConfigInit(ViperGetBool("force"))
	},
}
