package bluetooth

import (
	"errors"
	"testing"
)

func TestIsBlueUtilInstalled_Found(t *testing.T) {
	orig := lookPath
	defer func() { lookPath = orig }()

	lookPath = func(file string) (string, error) {
		if file == "blueutil" {
			return "/opt/homebrew/bin/blueutil", nil
		}
		return "", errors.New("not found")
	}

	if !IsBlueUtilInstalled() {
		t.Error("expected blueutil to be detected as installed")
	}
}

func TestIsBlueUtilInstalled_NotFound(t *testing.T) {
	orig := lookPath
	defer func() { lookPath = orig }()

	lookPath = func(file string) (string, error) {
		return "", errors.New("not found")
	}

	if IsBlueUtilInstalled() {
		t.Error("expected blueutil to be detected as not installed")
	}
}

func TestConnect_BlueUtilNotInstalled(t *testing.T) {
	origLook := lookPath
	defer func() { lookPath = origLook }()

	lookPath = func(file string) (string, error) {
		return "", errors.New("not found")
	}

	err := Connect("AA:BB:CC:DD:EE:FF")
	if !errors.Is(err, ErrBlueUtilNotInstalled) {
		t.Errorf("expected ErrBlueUtilNotInstalled, got %v", err)
	}
}

func TestConnect_Success(t *testing.T) {
	origLook := lookPath
	origCmd := commandRunner
	defer func() {
		lookPath = origLook
		commandRunner = origCmd
	}()

	lookPath = func(file string) (string, error) {
		return "/opt/homebrew/bin/blueutil", nil
	}

	var gotArgs []string
	commandRunner = func(name string, args ...string) ([]byte, error) {
		gotArgs = append([]string{name}, args...)
		return nil, nil
	}

	err := Connect("AA:BB:CC:DD:EE:FF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(gotArgs) != 3 || gotArgs[0] != "blueutil" || gotArgs[1] != "--connect" || gotArgs[2] != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("unexpected command args: %v", gotArgs)
	}
}

func TestConnect_Failure(t *testing.T) {
	origLook := lookPath
	origCmd := commandRunner
	defer func() {
		lookPath = origLook
		commandRunner = origCmd
	}()

	lookPath = func(file string) (string, error) {
		return "/opt/homebrew/bin/blueutil", nil
	}

	commandRunner = func(name string, args ...string) ([]byte, error) {
		return nil, errors.New("connection failed")
	}

	err := Connect("AA:BB:CC:DD:EE:FF")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, errors.Unwrap(err)) {
		// Just check it contains useful context
	}
}

func TestDisconnect_BlueUtilNotInstalled(t *testing.T) {
	origLook := lookPath
	defer func() { lookPath = origLook }()

	lookPath = func(file string) (string, error) {
		return "", errors.New("not found")
	}

	err := Disconnect("AA:BB:CC:DD:EE:FF")
	if !errors.Is(err, ErrBlueUtilNotInstalled) {
		t.Errorf("expected ErrBlueUtilNotInstalled, got %v", err)
	}
}

func TestDisconnect_Success(t *testing.T) {
	origLook := lookPath
	origCmd := commandRunner
	defer func() {
		lookPath = origLook
		commandRunner = origCmd
	}()

	lookPath = func(file string) (string, error) {
		return "/opt/homebrew/bin/blueutil", nil
	}

	var gotArgs []string
	commandRunner = func(name string, args ...string) ([]byte, error) {
		gotArgs = append([]string{name}, args...)
		return nil, nil
	}

	err := Disconnect("AA:BB:CC:DD:EE:FF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(gotArgs) != 3 || gotArgs[0] != "blueutil" || gotArgs[1] != "--disconnect" || gotArgs[2] != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("unexpected command args: %v", gotArgs)
	}
}

func TestRemove_BlueUtilNotInstalled(t *testing.T) {
	origLook := lookPath
	defer func() { lookPath = origLook }()

	lookPath = func(file string) (string, error) {
		return "", errors.New("not found")
	}

	err := Remove("AA:BB:CC:DD:EE:FF")
	if !errors.Is(err, ErrBlueUtilNotInstalled) {
		t.Errorf("expected ErrBlueUtilNotInstalled, got %v", err)
	}
}

func TestRemove_Success(t *testing.T) {
	origLook := lookPath
	origCmd := commandRunner
	defer func() {
		lookPath = origLook
		commandRunner = origCmd
	}()

	lookPath = func(file string) (string, error) {
		return "/opt/homebrew/bin/blueutil", nil
	}

	var gotArgs []string
	commandRunner = func(name string, args ...string) ([]byte, error) {
		gotArgs = append([]string{name}, args...)
		return nil, nil
	}

	err := Remove("AA:BB:CC:DD:EE:FF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(gotArgs) != 3 || gotArgs[0] != "blueutil" || gotArgs[1] != "--unpair" || gotArgs[2] != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("unexpected command args: %v", gotArgs)
	}
}

func TestPowerOn_Success(t *testing.T) {
	origCmd := commandRunner
	defer func() { commandRunner = origCmd }()

	var calls [][]string
	commandRunner = func(name string, args ...string) ([]byte, error) {
		calls = append(calls, append([]string{name}, args...))
		return nil, nil
	}

	err := PowerOn()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(calls))
	}

	// First call: defaults write
	if calls[0][0] != "defaults" || calls[0][5] != "1" {
		t.Errorf("unexpected first command: %v", calls[0])
	}

	// Second call: killall
	if calls[1][0] != "killall" || calls[1][1] != "-HUP" || calls[1][2] != "bluetoothd" {
		t.Errorf("unexpected second command: %v", calls[1])
	}
}

func TestPowerOff_Success(t *testing.T) {
	origCmd := commandRunner
	defer func() { commandRunner = origCmd }()

	var calls [][]string
	commandRunner = func(name string, args ...string) ([]byte, error) {
		calls = append(calls, append([]string{name}, args...))
		return nil, nil
	}

	err := PowerOff()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(calls))
	}

	// Check power state is 0
	if calls[0][5] != "0" {
		t.Errorf("expected power state 0, got %s", calls[0][5])
	}
}

func TestPowerOn_DefaultsWriteError(t *testing.T) {
	origCmd := commandRunner
	defer func() { commandRunner = origCmd }()

	commandRunner = func(name string, args ...string) ([]byte, error) {
		if name == "defaults" {
			return nil, errors.New("permission denied")
		}
		return nil, nil
	}

	err := PowerOn()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPowerOn_KillallError(t *testing.T) {
	origCmd := commandRunner
	defer func() { commandRunner = origCmd }()

	commandRunner = func(name string, args ...string) ([]byte, error) {
		if name == "killall" {
			return nil, errors.New("process not found")
		}
		return nil, nil
	}

	err := PowerOn()
	if err == nil {
		t.Fatal("expected error when killall fails")
	}
}
