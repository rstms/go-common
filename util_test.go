package util

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func initTestConfig(t *testing.T) {
	viper.SetConfigFile("testdata/config.yaml")
	err := viper.ReadInConfig()
	require.Nil(t, err)
}

func TestLog(t *testing.T) {
	OpenLog()
	log.Println("log message")
	CloseLog()
}
