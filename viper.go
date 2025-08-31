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
	viperKey := ViperKey(key)
	value := viper.GetBool(viperKey)
	if viper.GetBool(ViperKey("debug")) {
		log.Printf("ViperGetBool(%s) -> %s=%v\n", key, viperKey, value)
	}
	return value
}

func ViperGetString(key string) string {
	viperKey := ViperKey(key)
	value := Expand(viper.GetString(viperKey))
	if viper.GetBool(ViperKey("debug")) {
		log.Printf("ViperGetString(%s) -> %s=%v\n", key, viperKey, value)
	}
	return value
}

func ViperGetStringSlice(key string) []string {
	viperKey := ViperKey(key)
	values := []string{}
	for _, value := range viper.GetStringSlice(viperKey) {
		values = append(values, Expand(value))
	}
	if viper.GetBool(ViperKey("debug")) {
		log.Printf("ViperGetStringSlice(%s) -> %s=%v\n", key, viperKey, values)
	}
	return values
}

func ViperGetStringMapString(key string) map[string]string {
	viperKey := ViperKey(key)
	values := make(map[string]string)
	for key, value := range viper.GetStringMapString(viperKey) {
		values[key] = Expand(value)
	}
	if viper.GetBool(ViperKey("debug")) {
		log.Printf("ViperGetStringMapString(%s) -> %s=%v\n", key, viperKey, values)
	}
	return values
}

func ViperGetInt(key string) int {
	viperKey := ViperKey(key)
	value := viper.GetInt(viperKey)
	if viper.GetBool(ViperKey("debug")) {
		log.Printf("ViperGetInt(%s) -> %s=%d\n", key, viperKey, value)
	}
	return value
}

func ViperGetInt64(key string) int64 {
	viperKey := ViperKey(key)
	value := viper.GetInt64(viperKey)
	if viper.GetBool(ViperKey("debug")) {
		log.Printf("ViperGetInt64(%s) -> %s=%d\n", key, viperKey, value)
	}
	return value
}

func ViperSet(key string, value any) {
	viperKey := ViperKey(key)
	viper.Set(viperKey, value)
	if viper.GetBool(ViperKey("debug")) {
		log.Printf("ViperSet(%s) %s=%v\n", key, viperKey, value)
	}
}

func ViperSetDefault(key string, value any) {
	viperKey := ViperKey(key)
	viper.SetDefault(viperKey, value)
	if viper.GetBool(ViperKey("debug")) {
		log.Printf("ViperSetDefault(%s) %s=%v\n", key, viperKey, value)
	}
}
