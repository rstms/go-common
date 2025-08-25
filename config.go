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

var configFilename string

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
	name := rootCmd.Name()
	configFilename := viper.ConfigFileUsed()
	switch configFilename {
	case "":
		userConfig, err := os.UserConfigDir()
		cobra.CheckErr(err)
		configDir := filepath.Join(userConfig, name)
		err = os.MkdirAll(configDir, 0700)
		cobra.CheckErr(err)
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
	log.Println("initConfig")
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
	if viper.ConfigFileUsed() != "" && viper.GetBool("verbose") {
		log.Println("Using config file:", viper.ConfigFileUsed())
	}
}
