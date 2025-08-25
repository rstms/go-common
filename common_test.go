package common

import (
	"errors"
	"github.com/stretchr/testify/require"
	"log"
	"path/filepath"
	"testing"
)

func initTestConfig(t *testing.T) {
	Init("go-common", Version, filepath.Join("testdata", "config.yaml"))
	//viper.WriteConfigTo(os.Stdout)
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

func fatal(err error) error {
	return Fatal(err)
}

func fatalf(format string, args ...interface{}) error {
	return Fatalf(format, args...)
}

func TestFatal(t *testing.T) {
	expected := errors.New("expected_error")
	err := fatal(expected)
	log.Println(err)
	require.IsType(t, expected, err)
	err = fatalf("printf_type: %s: %v", "with_value", expected)
	require.IsType(t, expected, err)
	log.Println(err)
}

func TestWarning(t *testing.T) {
	Warning("string warning")
	Warning("formatted warning: %v", true)
}
