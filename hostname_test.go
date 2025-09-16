package common

import (
	"github.com/stretchr/testify/require"
	"log"
	"strings"
	"testing"
)

func TestHostShortname(t *testing.T) {
	host, err := HostShortname()
	require.Nil(t, err)
	require.NotEmpty(t, host)
	require.False(t, strings.Contains(host, "."))
	log.Printf("HostShortname: %s\n", host)
}

func TestHostDomain(t *testing.T) {
	domain, err := HostDomain()
	require.Nil(t, err)
	require.NotEmpty(t, domain)
	require.True(t, strings.Contains(domain, "."))
	log.Printf("HostDomain: %s\n", domain)
}

func TestHostFQDN(t *testing.T) {
	fqdn, err := HostFQDN()
	require.Nil(t, err)
	require.NotEmpty(t, fqdn)
	require.True(t, strings.Contains(fqdn, "."))
	log.Printf("HostFQDN: %s\n", fqdn)
}
