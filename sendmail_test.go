package common

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestSendmail(t *testing.T) {
	initTestConfig(t)

	// write password into temp file
	passwordFile, err := os.CreateTemp("", "test*")
	require.Nil(t, err)
	defer os.Remove(passwordFile.Name())
	_, err = passwordFile.Write([]byte(os.Getenv("SMTP_PASSWORD")))
	require.Nil(t, err)
	err = passwordFile.Close()
	require.Nil(t, err)
	ViperSet("smtp.password", "@"+passwordFile.Name())

	// initialize the sendmail object
	s, err := NewSendmail(
		ViperGetString("smtp.hostname"),
		ViperGetInt("smtp.port"),
		ViperGetString("smtp.username"),
		ViperGetString("smtp.password"),
		ViperGetString("smtp.cafile"),
	)
	require.Nil(t, err)

	// setup message parameters
	to := ViperGetString("test.sendmail.to")
	from := ViperGetString("test.sendmail.from")
	subject := "common.Sendmail test message"
	body := []byte("howdy, howdy, howdy\nthis is a test message\n")

	// send the mail
	err = s.Send(to, from, subject, body)
	require.Nil(t, err)
}
