//go:build ignore
// +build ignore

package main

import (
	"bufio"
	"os"

	"github.com/Pluslab/cyphonic/noded/internal/logger"
	"github.com/joho/godotenv"
)

// EnvLoad loads .env file.
func EnvLoad() {
	err := godotenv.Load()
	if err != nil {
		logger.LogErr("Error loading .env file", "error", err)
	}
}

func main() {
	EnvLoad()

	var resolv []byte

	// FIXME: パスのマジックストリング
	fp, err := os.Open("/etc/resolv.conf")
	if err != nil {
		logger.LogErr("Can't open /etc/resolv.conf", "error", err)
	}
	defer func() {
		if err := fp.Close(); err != nil {
			logger.LogErr("Can't close /etc/resolv.conf", "error", err)
		}
	}()

	scanner := bufio.NewScanner(fp)

	for scanner.Scan() {
		if scanner.Text() != ("nameserver "+os.Getenv("TUN_DNS_INTERFACE_IPv6")) || scanner.Text() != ("nameserver "+os.Getenv("TUN_DNS_INTERFACE_IPv4")) {
			resolv = append(resolv, []byte(scanner.Text()+"\n")...)
		}
	}

	if err = os.WriteFile("/etc/resolv.conf", resolv, 0o644); err != nil {
		logger.LogErr("Can't write /etc/resolv.conf", "error", err)
	}
}
