package bluetooth

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Device represents a paired Bluetooth device.
type Device struct {
	Name         string `json:"name"`
	Address      string `json:"address"`
	MinorType    string `json:"minor_type"`
	Connected    bool   `json:"connected"`
	BatteryLevel int    `json:"battery_level"` // 0-100, or -1 if unknown
	RSSI         int    `json:"rssi"`          // signal strength, or 0 if unknown
}

// systemProfilerOutput represents the top-level system_profiler JSON.
type systemProfilerOutput struct {
	SPBluetoothDataType []bluetoothData `json:"SPBluetoothDataType"`
}

// bluetoothData represents the Bluetooth data section.
type bluetoothData struct {
	ControllerProperties map[string]string        `json:"controller_properties"`
	DeviceConnected      []map[string]interface{} `json:"device_connected"`
	DeviceNotConnected   []map[string]interface{} `json:"device_not_connected"`
}

// commandRunner abstracts command execution for testing.
var commandRunner = func(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

// ListDevices returns all paired Bluetooth devices (connected and disconnected).
func ListDevices() ([]Device, error) {
	out, err := commandRunner("system_profiler", "SPBluetoothDataType", "-json")
	if err != nil {
		return nil, fmt.Errorf("failed to run system_profiler: %w", err)
	}
	return ParseDevices(out)
}

// ParseDevices parses system_profiler SPBluetoothDataType JSON output into devices.
func ParseDevices(data []byte) ([]Device, error) {
	var sp systemProfilerOutput
	if err := json.Unmarshal(data, &sp); err != nil {
		return nil, fmt.Errorf("failed to parse bluetooth data: %w", err)
	}

	if len(sp.SPBluetoothDataType) == 0 {
		return nil, nil
	}

	bt := sp.SPBluetoothDataType[0]
	var devices []Device

	for _, entry := range bt.DeviceConnected {
		d := parseDeviceEntry(entry, true)
		devices = append(devices, d...)
	}

	for _, entry := range bt.DeviceNotConnected {
		d := parseDeviceEntry(entry, false)
		devices = append(devices, d...)
	}

	return devices, nil
}

// parseDeviceEntry parses a single device entry from the system_profiler JSON.
// Each entry is a map with one key (device name) mapping to its properties.
func parseDeviceEntry(entry map[string]interface{}, connected bool) []Device {
	var devices []Device
	for name, propsRaw := range entry {
		props, ok := propsRaw.(map[string]interface{})
		if !ok {
			continue
		}
		d := Device{
			Name:         name,
			Address:      getString(props, "device_address"),
			MinorType:    getString(props, "device_minorType"),
			Connected:    connected,
			BatteryLevel: parseBatteryLevel(props),
			RSSI:         parseRSSI(props),
		}
		devices = append(devices, d)
	}
	return devices
}

// parseBatteryLevel extracts battery level from device properties.
// It checks device_batteryLevel, device_batteryLevelMain, device_batteryLevelLeft,
// and device_batteryLevelRight fields. Returns the first found, or -1 if unknown.
func parseBatteryLevel(props map[string]interface{}) int {
	// Try main battery level fields in priority order
	for _, key := range []string{
		"device_batteryLevel",
		"device_batteryLevelMain",
		"device_batteryLevelLeft",
		"device_batteryLevelCase",
	} {
		if val, ok := props[key]; ok {
			if level := parseBatteryString(val); level >= 0 {
				return level
			}
		}
	}
	return -1
}

// parseBatteryString parses a battery level string like "75%" to an integer.
func parseBatteryString(val interface{}) int {
	s, ok := val.(string)
	if !ok {
		return -1
	}
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "%")
	s = strings.TrimSpace(s)
	level, err := strconv.Atoi(s)
	if err != nil {
		return -1
	}
	if level < 0 || level > 100 {
		return -1
	}
	return level
}

// parseRSSI extracts RSSI value from device properties.
func parseRSSI(props map[string]interface{}) int {
	val, ok := props["device_rssi"]
	if !ok {
		return 0
	}
	s, ok := val.(string)
	if !ok {
		return 0
	}
	rssi, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return rssi
}

// getString safely gets a string value from a map.
func getString(m map[string]interface{}, key string) string {
	val, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := val.(string)
	if !ok {
		return ""
	}
	return s
}

// ListConnected returns only connected Bluetooth devices.
func ListConnected() ([]Device, error) {
	devices, err := ListDevices()
	if err != nil {
		return nil, err
	}
	var connected []Device
	for _, d := range devices {
		if d.Connected {
			connected = append(connected, d)
		}
	}
	return connected, nil
}

// GetDevice finds a device by name or address (case-insensitive).
func GetDevice(nameOrAddr string) (*Device, error) {
	devices, err := ListDevices()
	if err != nil {
		return nil, err
	}
	search := strings.ToLower(nameOrAddr)
	for _, d := range devices {
		if strings.ToLower(d.Name) == search || strings.ToLower(d.Address) == search {
			return &d, nil
		}
	}
	return nil, fmt.Errorf("device not found: %s", nameOrAddr)
}
