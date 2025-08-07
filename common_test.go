package common

import (
	"errors"
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

func TestHexDump(t *testing.T) {
	data := []byte("howdy\nhowdy\nhowdy\n")
	log.Println("label\n" + HexDump(data))
}

func TestFatal(t *testing.T) {
	err := Fatal("string message")
	expected := errors.New("expected")
	log.Println(err)
	require.IsType(t, expected, err)
	err = Fatal("printf error: %s", "with_value")
	require.IsType(t, expected, err)
	log.Println(err)
}

func TestWarning(t *testing.T) {
	Warning("string warning")
	Warning("formatted warning: %v", true)
}
