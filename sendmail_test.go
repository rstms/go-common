package common

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestSendmail(t *testing.T) {
	initTestConfig(t)
	passwordFile, err := os.CreateTemp("", "test*")
	require.Nil(t, err)
	defer os.Remove(passwordFile.Name())
	_, err = passwordFile.Write([]byte(os.Getenv("SMTP_PASSWORD")))
	require.Nil(t, err)
	err = passwordFile.Close()
	require.Nil(t, err)
	ViperSet("smtp.password_file", passwordFile.Name())
	to := ViperGetString("test.sendmail.to")
	from := ViperGetString("test.sendmail.from")
	subject := "common.Sendmail test message"
	body := []byte("howdy, howdy, howdy")
	err = Sendmail(to, from, subject, body, nil)
	require.Nil(t, err)
}
