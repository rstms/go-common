package common

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func initTestConfig(t *testing.T) {
	name := "go-common"
	Init(&name, &Version)
	viper.SetConfigFile("testdata/config.yaml")
	err := viper.ReadInConfig()
	require.Nil(t, err)
}

func TestViperGet(t *testing.T) {
	initTestConfig(t)
	testValue := ViperGetString("test_value")
	require.Equal(t, "testing123", testValue)
}

func TestLog(t *testing.T) {
	initTestConfig(t)
	OpenLog()
	log.Println("log message")
	CloseLog()
}

func TestDebugLog(t *testing.T) {
	initTestConfig(t)
	ViperSet("debug", true)
	OpenLog()
	log.Println("log message")
	CloseLog()
}
