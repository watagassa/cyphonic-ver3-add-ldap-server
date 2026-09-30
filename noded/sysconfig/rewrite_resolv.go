package sysconfig

import (
	"bufio"
	"os"

	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
)

// RewriteResolv rewrites /etc/resolv.conf.
func RewriteResolv(virtualIPType int, tunDnsInterfaceIPv4, tunDnsInterfaceIPv6 string) {
	// FIXME: パスのマジックストリング
	fp, err := os.Open("/etc/resolv.conf")
	if err != nil {
		config.LogErr("Can't open /etc/resolv.conf", "error", err)
	}
	defer func() {
		if err := fp.Close(); err != nil {
			config.LogErr("Can't close /etc/resolv.conf", "error", err)
		}
	}()

	scanner := bufio.NewScanner(fp)
	var resolv []byte

	switch virtualIPType {
	case 4:
		resolv = []byte("nameserver " + tunDnsInterfaceIPv4 + "\n")
		for scanner.Scan() {
			if scanner.Text() != ("nameserver " + tunDnsInterfaceIPv4) {
				resolv = append(resolv, []byte(scanner.Text()+"\n")...)
			}
		}
	case 6:
		resolv = []byte("nameserver " + tunDnsInterfaceIPv6 + "\n")
		for i := 1; scanner.Scan(); i++ {
			if scanner.Text() != ("nameserver " + tunDnsInterfaceIPv6) {
				resolv = append(resolv, []byte(scanner.Text()+"\n")...)
			}
		}
	}

	if err = os.WriteFile("/etc/resolv.conf", resolv, 0o644); err != nil {
		config.LogErr("Can't write /etc/resolv.conf", "error", err)
	}
}
