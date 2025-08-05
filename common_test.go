package common

import (
	"github.com/stretchr/testify/require"
	"log"
	"path/filepath"
	"testing"
)

func initTestConfig(t *testing.T) {
	Init("go-common", Version)
	InitConfig(filepath.Join("testdata", "config.yaml"))
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
