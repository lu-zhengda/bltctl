package bluetooth

import (
	"encoding/json"
	"fmt"
	"strings"
)

// DiagReport contains Bluetooth diagnostic information.
type DiagReport struct {
	PowerState       string            `json:"power_state"`
	ControllerInfo   map[string]string `json:"controller_info"`
	ConnectedDevices []Device          `json:"connected_devices"`
	RecentErrors     []string          `json:"recent_errors"`
}

// Diagnose performs a comprehensive Bluetooth diagnostic check.
func Diagnose() (*DiagReport, error) {
	report := &DiagReport{
		ControllerInfo: make(map[string]string),
	}

	// Get power state from system_profiler
	spOut, err := commandRunner("system_profiler", "SPBluetoothDataType", "-json")
	if err != nil {
		return nil, fmt.Errorf("failed to run system_profiler: %w", err)
	}

	if err := parseDiagnosticData(spOut, report); err != nil {
		return nil, err
	}

	// Get connected devices
	devices, err := ParseDevices(spOut)
	if err != nil {
		return nil, fmt.Errorf("failed to parse devices: %w", err)
	}
	for _, d := range devices {
		if d.Connected {
			report.ConnectedDevices = append(report.ConnectedDevices, d)
		}
	}

	// Get recent Bluetooth errors from log
	logOut, err := commandRunner("log", "show",
		"--predicate", `subsystem == "com.apple.bluetooth"`,
		"--last", "5m",
		"--style", "compact")
	if err != nil {
		// Log collection is best-effort; don't fail the whole report
		report.RecentErrors = append(report.RecentErrors, fmt.Sprintf("could not collect logs: %v", err))
	} else {
		report.RecentErrors = parseLogErrors(string(logOut))
	}

	return report, nil
}

// parseDiagnosticData extracts controller info and power state from system_profiler JSON.
func parseDiagnosticData(data []byte, report *DiagReport) error {
	var sp systemProfilerOutput
	if err := json.Unmarshal(data, &sp); err != nil {
		return fmt.Errorf("failed to parse bluetooth data: %w", err)
	}

	if len(sp.SPBluetoothDataType) == 0 {
		report.PowerState = "unknown"
		return nil
	}

	bt := sp.SPBluetoothDataType[0]
	for k, v := range bt.ControllerProperties {
		report.ControllerInfo[k] = v
	}

	// Extract power state from controller_state
	if state, ok := bt.ControllerProperties["controller_state"]; ok {
		switch state {
		case "attrib_on":
			report.PowerState = "on"
		case "attrib_off":
			report.PowerState = "off"
		default:
			report.PowerState = state
		}
	} else {
		report.PowerState = "unknown"
	}

	return nil
}

// parseLogErrors extracts error lines from Bluetooth log output.
func parseLogErrors(logOutput string) []string {
	var errors []string
	lines := strings.Split(logOutput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if strings.Contains(lower, "error") ||
			strings.Contains(lower, "fail") ||
			strings.Contains(lower, "disconnect") ||
			strings.Contains(lower, "timeout") {
			errors = append(errors, line)
		}
	}
	return errors
}

// Reset kills the Bluetooth daemon, which macOS auto-restarts.
// Requires sudo; returns a clear error if not root.
func Reset() error {
	_, err := commandRunner("sudo", "pkill", "bluetoothd")
	if err != nil {
		return fmt.Errorf("failed to reset bluetooth (sudo required): %w", err)
	}
	return nil
}
