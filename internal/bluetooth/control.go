package bluetooth

import (
	"errors"
	"fmt"
	"os/exec"
)

// ErrBlueUtilNotInstalled is returned when blueutil is required but not available.
var ErrBlueUtilNotInstalled = errors.New("blueutil required — brew install blueutil")

// ErrSudoRequired is returned when a command requires root privileges.
var ErrSudoRequired = errors.New("this operation requires sudo")

// IsBlueUtilInstalled checks if blueutil is available on the system.
func IsBlueUtilInstalled() bool {
	_, err := lookPath("blueutil")
	return err == nil
}

// lookPath abstracts exec.LookPath for testing.
var lookPath = exec.LookPath

// PowerOn enables the Bluetooth controller.
// Requires sudo; returns a clear error if not root.
func PowerOn() error {
	return setPowerState(1)
}

// PowerOff disables the Bluetooth controller.
// Requires sudo; returns a clear error if not root.
func PowerOff() error {
	return setPowerState(0)
}

// setPowerState sets the Bluetooth controller power state.
func setPowerState(state int) error {
	stateStr := fmt.Sprintf("%d", state)
	_, err := commandRunner("defaults", "write",
		"/Library/Preferences/com.apple.Bluetooth",
		"ControllerPowerState", "-int", stateStr)
	if err != nil {
		return fmt.Errorf("failed to set power state (sudo required): %w", err)
	}

	_, err = commandRunner("killall", "-HUP", "bluetoothd")
	if err != nil {
		return fmt.Errorf("failed to restart bluetoothd: %w", err)
	}

	return nil
}

// requireBlueUtil returns an error if blueutil is not installed.
func requireBlueUtil() error {
	if !IsBlueUtilInstalled() {
		return ErrBlueUtilNotInstalled
	}
	return nil
}

// Connect connects to a device by address using blueutil.
func Connect(address string) error {
	if err := requireBlueUtil(); err != nil {
		return err
	}
	_, err := commandRunner("blueutil", "--connect", address)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", address, err)
	}
	return nil
}

// Disconnect disconnects a device by address using blueutil.
func Disconnect(address string) error {
	if err := requireBlueUtil(); err != nil {
		return err
	}
	_, err := commandRunner("blueutil", "--disconnect", address)
	if err != nil {
		return fmt.Errorf("failed to disconnect %s: %w", address, err)
	}
	return nil
}

// Remove unpairs a device by address using blueutil.
func Remove(address string) error {
	if err := requireBlueUtil(); err != nil {
		return err
	}
	_, err := commandRunner("blueutil", "--unpair", address)
	if err != nil {
		return fmt.Errorf("failed to remove %s: %w", address, err)
	}
	return nil
}
