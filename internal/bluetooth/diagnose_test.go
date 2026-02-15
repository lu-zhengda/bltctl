package bluetooth

import (
	"errors"
	"testing"
)

const diagJSON = `{
  "SPBluetoothDataType" : [
    {
      "controller_properties" : {
        "controller_address" : "BC:D0:74:22:43:D6",
        "controller_chipset" : "BCM_4387",
        "controller_state" : "attrib_on",
        "controller_firmwareVersion" : "23.1.623.4111",
        "controller_transport" : "PCIe"
      },
      "device_connected" : [
        {
          "AirPods Max" : {
            "device_address" : "70:F9:4A:7A:8B:CA",
            "device_minorType" : "Headphones",
            "device_batteryLevel" : "85%"
          }
        }
      ],
      "device_not_connected" : [
        {
          "AirPods Pro" : {
            "device_address" : "74:15:F5:4E:D0:50",
            "device_minorType" : "Headphones"
          }
        }
      ]
    }
  ]
}`

const diagOffJSON = `{
  "SPBluetoothDataType" : [
    {
      "controller_properties" : {
        "controller_state" : "attrib_off"
      }
    }
  ]
}`

func TestParseDiagnosticData_PowerOn(t *testing.T) {
	report := &DiagReport{ControllerInfo: make(map[string]string)}
	err := parseDiagnosticData([]byte(diagJSON), report)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.PowerState != "on" {
		t.Errorf("expected power state 'on', got '%s'", report.PowerState)
	}
	if report.ControllerInfo["controller_chipset"] != "BCM_4387" {
		t.Errorf("expected chipset BCM_4387, got %s", report.ControllerInfo["controller_chipset"])
	}
	if report.ControllerInfo["controller_transport"] != "PCIe" {
		t.Errorf("expected transport PCIe, got %s", report.ControllerInfo["controller_transport"])
	}
}

func TestParseDiagnosticData_PowerOff(t *testing.T) {
	report := &DiagReport{ControllerInfo: make(map[string]string)}
	err := parseDiagnosticData([]byte(diagOffJSON), report)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.PowerState != "off" {
		t.Errorf("expected power state 'off', got '%s'", report.PowerState)
	}
}

func TestParseDiagnosticData_Empty(t *testing.T) {
	report := &DiagReport{ControllerInfo: make(map[string]string)}
	err := parseDiagnosticData([]byte(noDataJSON), report)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.PowerState != "unknown" {
		t.Errorf("expected power state 'unknown', got '%s'", report.PowerState)
	}
}

func TestParseDiagnosticData_InvalidJSON(t *testing.T) {
	report := &DiagReport{ControllerInfo: make(map[string]string)}
	err := parseDiagnosticData([]byte("not json"), report)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseLogErrors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			"mixed lines",
			"2024-01-01 normal operation\n2024-01-01 error: connection refused\n2024-01-01 device connected\n",
			1,
		},
		{
			"error and fail",
			"error: something\nfailed to connect\ntimeout waiting\nnormal line\ndisconnected\n",
			4,
		},
		{
			"empty",
			"",
			0,
		},
		{
			"no errors",
			"device paired\nconnection established\nservice discovered\n",
			0,
		},
		{
			"case insensitive",
			"ERROR: big problem\nFailed operation\nTIMEOUT expired\n",
			3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLogErrors(tt.input)
			if len(got) != tt.expected {
				t.Errorf("expected %d errors, got %d: %v", tt.expected, len(got), got)
			}
		})
	}
}

func TestDiagnose_Success(t *testing.T) {
	origCmd := commandRunner
	defer func() { commandRunner = origCmd }()

	callCount := 0
	commandRunner = func(name string, args ...string) ([]byte, error) {
		callCount++
		if name == "system_profiler" {
			return []byte(diagJSON), nil
		}
		if name == "log" {
			return []byte("2024-01-01 error: test error\n2024-01-01 normal line\n"), nil
		}
		return nil, nil
	}

	report, err := Diagnose()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.PowerState != "on" {
		t.Errorf("expected power state 'on', got '%s'", report.PowerState)
	}

	if len(report.ConnectedDevices) != 1 {
		t.Errorf("expected 1 connected device, got %d", len(report.ConnectedDevices))
	}

	if len(report.RecentErrors) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(report.RecentErrors), report.RecentErrors)
	}
}

func TestDiagnose_LogError(t *testing.T) {
	origCmd := commandRunner
	defer func() { commandRunner = origCmd }()

	commandRunner = func(name string, args ...string) ([]byte, error) {
		if name == "system_profiler" {
			return []byte(diagJSON), nil
		}
		if name == "log" {
			return nil, errors.New("permission denied")
		}
		return nil, nil
	}

	report, err := Diagnose()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still succeed but include the log error in RecentErrors
	if len(report.RecentErrors) != 1 {
		t.Fatalf("expected 1 error from log failure, got %d", len(report.RecentErrors))
	}
	if report.RecentErrors[0] == "" {
		t.Error("expected non-empty error message")
	}
}

func TestDiagnose_SystemProfilerError(t *testing.T) {
	origCmd := commandRunner
	defer func() { commandRunner = origCmd }()

	commandRunner = func(name string, args ...string) ([]byte, error) {
		if name == "system_profiler" {
			return nil, errors.New("command failed")
		}
		return nil, nil
	}

	_, err := Diagnose()
	if err == nil {
		t.Fatal("expected error when system_profiler fails")
	}
}

func TestReset_Success(t *testing.T) {
	origCmd := commandRunner
	defer func() { commandRunner = origCmd }()

	var gotArgs []string
	commandRunner = func(name string, args ...string) ([]byte, error) {
		gotArgs = append([]string{name}, args...)
		return nil, nil
	}

	err := Reset()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(gotArgs) != 3 || gotArgs[0] != "sudo" || gotArgs[1] != "pkill" || gotArgs[2] != "bluetoothd" {
		t.Errorf("unexpected command args: %v", gotArgs)
	}
}

func TestReset_Error(t *testing.T) {
	origCmd := commandRunner
	defer func() { commandRunner = origCmd }()

	commandRunner = func(name string, args ...string) ([]byte, error) {
		return nil, errors.New("permission denied")
	}

	err := Reset()
	if err == nil {
		t.Fatal("expected error when reset fails")
	}
}
