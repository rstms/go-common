package common

import (
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
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

func ViperKey(key string) string {
	var prefix string
	name := ProgramName()
	if name != "" {
		prefix = name + "."
	}
	ret := strings.ToLower(strings.ReplaceAll(prefix+key, "-", "_"))
	return ret
}

func ViperGetBool(key string) bool {
	return viper.GetBool(ViperKey(key))
}

func ViperGetString(key string) string {
	return Expand(viper.GetString(ViperKey(key)))
}

func ViperGetStringSlice(key string) []string {
	values := []string{}
	for _, value := range viper.GetStringSlice(ViperKey(key)) {
		values = append(values, Expand(value))
	}
	return values
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
