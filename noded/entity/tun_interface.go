package entity

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/songgao/water"
)

type TunInterface struct {
	tun     *water.Interface
	address string
}

// NewTunInterface returns Tunnel device.
func NewTunInterface(name string, address string, prefix int) (*TunInterface, error) {
	addr := address + "/" + strconv.Itoa(prefix)

	switch runtime.GOOS {
	case "linux":
		config := water.Config{
			DeviceType: water.TUN,
		}
		config.Name = name

		ifce, err := water.New(config)
		if err != nil {
			return nil, fmt.Errorf("unable to create TUN/TAP interface: %w", err)
		}

		iface := &TunInterface{
			tun:     ifce,
			address: addr,
		}

		return iface, nil

	case "darwin":
		ifce, err := water.New(water.Config{
			DeviceType: water.TUN,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create TUN/TAP interface: %w", err)
		}

		iface := &TunInterface{
			tun:     ifce,
			address: addr,
		}

		return iface, nil

	case "windows":
		return nil, fmt.Errorf("windows is not supported")

	default:
		return nil, fmt.Errorf("%s is not supported", runtime.GOOS)
	}
}

func (iface *TunInterface) Up() error {
	switch runtime.GOOS {
	case "linux":
		if err := execCmd("ip", []string{"addr", "add", iface.address, "dev", iface.tun.Name()}); err != nil {
			return fmt.Errorf("ip command add fail: %w", err)
		}

		// TODO: Set any value for MTU.
		if err := execCmd("ip", []string{"link", "set", "dev", iface.tun.Name(), "up", "mtu", "1368"}); err != nil {
			return fmt.Errorf("ip command set fail: %w", err)
		}

	default:
		return fmt.Errorf("unsupported: %s %s", runtime.GOOS, runtime.GOARCH)
	}

	return nil
}

// Read function read the virtual interface.
func (iface *TunInterface) Read(buf []byte) (int, error) {
	n, err := iface.tun.Read(buf)
	if err != nil {
		return 0, fmt.Errorf("failed to read virtual interface: %w", err)
	}

	return n, nil
}

// Write function write the virtual interface.
func (iface *TunInterface) Write(buf []byte) (int, error) {
	return iface.tun.Write(buf)
}

// Close function closes the virtual interface.
func (iface *TunInterface) Close() error {
	if err := iface.tun.Close(); err != nil {
		return fmt.Errorf("failed to close virtual interface: %w", err)
	}

	return nil
}

func (iface *TunInterface) Name() string {
	return iface.tun.Name()
}

func (iface *TunInterface) Address() string {
	return iface.address
}

// execCmd executes the named program with the given arguments.
func execCmd(cmd string, args []string) error {
	execCmd := exec.Command(cmd, args...)
	if err := execCmd.Run(); err != nil {
		return fmt.Errorf("unable to execute command: %w", err)
	}

	return nil
}
