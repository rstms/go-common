package common

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSendmail(t *testing.T) {
	initTestConfig(t)

	to := ViperGetString("test.sendmail.to")
	from := ViperGetString("test.sendmail.from")
	subject := "common.Sendmail test message"
	body := []byte("howdy, howdy, howdy")

	err := Sendmail(to, from, subject, body, nil)
	require.Nil(t, err)
}
