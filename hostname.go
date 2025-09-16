package common

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func getHostnameDetail() (string, string, string, error) {
	var host string
	var domain string
	fqdn, err := os.Hostname()
	if err != nil {
		return "", "", "", Fatal(err)
	}
	if !strings.Contains(fqdn, ".") {
		host = fqdn
		fqdn = ""
		switch runtime.GOOS {
		case "openbsd":
			if IsFile("/etc/myname") {
				data, err := os.ReadFile("/etc/myname")
				if err != nil {
					return "", "", "", Fatal(err)
				}
				fqdn = strings.TrimSpace(string(data))
			}
		case "windows":
			cmd := exec.Command("ipconfig", "/all")
			var obuf bytes.Buffer
			cmd.Stdout = &obuf
			err := cmd.Run()
			if err != nil {
				return "", "", "", Fatal(err)
			}
			scanner := bufio.NewScanner(&obuf)
			for scanner.Scan() {
				line := scanner.Text()
				if fqdn == "" && strings.Contains(line, "Connection-specific DNS Suffix") {
					_, domain, ok := strings.Cut(line, ":")
					if ok {
						domain = strings.TrimSpace(domain)
						if domain != "" {
							fqdn = host + "." + strings.TrimSpace(domain)
						}
					}
				}
			}
			err = scanner.Err()
			if err != nil {
				return "", "", "", Fatal(err)
			}
		}
	}
	parts := strings.Split(fqdn, ".")
	if len(parts) < 2 {
		return "", "", "", Fatalf("unexpected fqdn format: %s", fqdn)
	}
	host = parts[0]
	domain = strings.Join(parts[1:], ".")
	return host, domain, fqdn, nil
}

func HostShortname() (string, error) {
	host, _, _, err := getHostnameDetail()
	return host, err
}

func HostDomain() (string, error) {
	_, domain, _, err := getHostnameDetail()
	return domain, err
}

func HostFQDN() (string, error) {
	_, _, fqdn, err := getHostnameDetail()
	return fqdn, err
}
