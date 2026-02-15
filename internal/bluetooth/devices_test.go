package bluetooth

import (
	"testing"
)

const sampleJSON = `{
  "SPBluetoothDataType" : [
    {
      "controller_properties" : {
        "controller_address" : "BC:D0:74:22:43:D6",
        "controller_chipset" : "BCM_4387",
        "controller_state" : "attrib_on"
      },
      "device_connected" : [
        {
          "AirPods Max" : {
            "device_address" : "70:F9:4A:7A:8B:CA",
            "device_firmwareVersion" : "7E108",
            "device_minorType" : "Headphones",
            "device_productID" : "0x201F",
            "device_vendorID" : "0x004C",
            "device_batteryLevel" : "85%"
          }
        }
      ],
      "device_not_connected" : [
        {
          "AirPods Pro" : {
            "device_address" : "74:15:F5:4E:D0:50",
            "device_batteryLevelCase" : "77%",
            "device_batteryLevelLeft" : "100%",
            "device_batteryLevelRight" : "100%",
            "device_minorType" : "Headphones",
            "device_productID" : "0x2014",
            "device_vendorID" : "0x004C"
          }
        },
        {
          "Beats Flex" : {
            "device_address" : "A8:91:3D:DE:91:C6",
            "device_minorType" : "Headphones",
            "device_productID" : "0x2010",
            "device_vendorID" : "0x004C"
          }
        },
        {
          "Home Theater" : {
            "device_address" : "94:EA:32:75:76:64",
            "device_rssi" : "-72"
          }
        },
        {
          "iPhone" : {
            "device_address" : "A8:8F:D9:7A:C8:14",
            "device_rssi" : "-42"
          }
        }
      ]
    }
  ]
}`

const emptyJSON = `{
  "SPBluetoothDataType" : [
    {
      "controller_properties" : {
        "controller_address" : "BC:D0:74:22:43:D6",
        "controller_state" : "attrib_on"
      }
    }
  ]
}`

const noDataJSON = `{
  "SPBluetoothDataType" : []
}`

func TestParseDevices(t *testing.T) {
	devices, err := ParseDevices([]byte(sampleJSON))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(devices) != 5 {
		t.Fatalf("expected 5 devices, got %d", len(devices))
	}

	// Verify connected device
	var airpodsMax *Device
	for i := range devices {
		if devices[i].Name == "AirPods Max" {
			airpodsMax = &devices[i]
			break
		}
	}
	if airpodsMax == nil {
		t.Fatal("AirPods Max not found")
	}
	if !airpodsMax.Connected {
		t.Error("AirPods Max should be connected")
	}
	if airpodsMax.Address != "70:F9:4A:7A:8B:CA" {
		t.Errorf("expected address 70:F9:4A:7A:8B:CA, got %s", airpodsMax.Address)
	}
	if airpodsMax.MinorType != "Headphones" {
		t.Errorf("expected MinorType Headphones, got %s", airpodsMax.MinorType)
	}
	if airpodsMax.BatteryLevel != 85 {
		t.Errorf("expected BatteryLevel 85, got %d", airpodsMax.BatteryLevel)
	}

	// Verify disconnected device with battery (from case level)
	var airpodsPro *Device
	for i := range devices {
		if devices[i].Name == "AirPods Pro" {
			airpodsPro = &devices[i]
			break
		}
	}
	if airpodsPro == nil {
		t.Fatal("AirPods Pro not found")
	}
	if airpodsPro.Connected {
		t.Error("AirPods Pro should be disconnected")
	}
	// batteryLevelLeft is checked before batteryLevelCase in priority order
	if airpodsPro.BatteryLevel != 100 {
		t.Errorf("expected BatteryLevel 100 (from Left), got %d", airpodsPro.BatteryLevel)
	}

	// Verify device without battery
	var beatsFlex *Device
	for i := range devices {
		if devices[i].Name == "Beats Flex" {
			beatsFlex = &devices[i]
			break
		}
	}
	if beatsFlex == nil {
		t.Fatal("Beats Flex not found")
	}
	if beatsFlex.BatteryLevel != -1 {
		t.Errorf("expected BatteryLevel -1 (unknown), got %d", beatsFlex.BatteryLevel)
	}

	// Verify device with RSSI
	var homeTheater *Device
	for i := range devices {
		if devices[i].Name == "Home Theater" {
			homeTheater = &devices[i]
			break
		}
	}
	if homeTheater == nil {
		t.Fatal("Home Theater not found")
	}
	if homeTheater.RSSI != -72 {
		t.Errorf("expected RSSI -72, got %d", homeTheater.RSSI)
	}
	if homeTheater.MinorType != "" {
		t.Errorf("expected empty MinorType, got %s", homeTheater.MinorType)
	}
}

func TestParseDevices_Empty(t *testing.T) {
	devices, err := ParseDevices([]byte(emptyJSON))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(devices))
	}
}

func TestParseDevices_NoData(t *testing.T) {
	devices, err := ParseDevices([]byte(noDataJSON))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if devices != nil {
		t.Errorf("expected nil devices, got %v", devices)
	}
}

func TestParseDevices_InvalidJSON(t *testing.T) {
	_, err := ParseDevices([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseDevices_ConnectedCount(t *testing.T) {
	devices, err := ParseDevices([]byte(sampleJSON))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	connected := 0
	for _, d := range devices {
		if d.Connected {
			connected++
		}
	}
	if connected != 1 {
		t.Errorf("expected 1 connected device, got %d", connected)
	}
}

func TestParseBatteryString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{"percentage", "85%", 85},
		{"zero", "0%", 0},
		{"hundred", "100%", 100},
		{"no percent sign", "50", 50},
		{"with spaces", " 75% ", 75},
		{"empty string", "", -1},
		{"non-numeric", "abc", -1},
		{"negative", "-5%", -1},
		{"over 100", "150%", -1},
		{"nil", nil, -1},
		{"number type", 42, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseBatteryString(tt.input)
			if got != tt.expected {
				t.Errorf("parseBatteryString(%v) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseRSSI(t *testing.T) {
	tests := []struct {
		name     string
		props    map[string]interface{}
		expected int
	}{
		{"present", map[string]interface{}{"device_rssi": "-45"}, -45},
		{"missing", map[string]interface{}{}, 0},
		{"non-string", map[string]interface{}{"device_rssi": 42}, 0},
		{"invalid", map[string]interface{}{"device_rssi": "abc"}, 0},
		{"positive", map[string]interface{}{"device_rssi": "10"}, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRSSI(tt.props)
			if got != tt.expected {
				t.Errorf("parseRSSI() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestListDevices_WithMock(t *testing.T) {
	orig := commandRunner
	defer func() { commandRunner = orig }()

	commandRunner = func(name string, args ...string) ([]byte, error) {
		return []byte(sampleJSON), nil
	}

	devices, err := ListDevices()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 5 {
		t.Errorf("expected 5 devices, got %d", len(devices))
	}
}

func TestListConnected_WithMock(t *testing.T) {
	orig := commandRunner
	defer func() { commandRunner = orig }()

	commandRunner = func(name string, args ...string) ([]byte, error) {
		return []byte(sampleJSON), nil
	}

	devices, err := ListConnected()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 1 {
		t.Errorf("expected 1 connected device, got %d", len(devices))
	}
	if devices[0].Name != "AirPods Max" {
		t.Errorf("expected AirPods Max, got %s", devices[0].Name)
	}
}

func TestGetDevice_ByName(t *testing.T) {
	orig := commandRunner
	defer func() { commandRunner = orig }()

	commandRunner = func(name string, args ...string) ([]byte, error) {
		return []byte(sampleJSON), nil
	}

	d, err := GetDevice("airpods max")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Name != "AirPods Max" {
		t.Errorf("expected AirPods Max, got %s", d.Name)
	}
}

func TestGetDevice_ByAddress(t *testing.T) {
	orig := commandRunner
	defer func() { commandRunner = orig }()

	commandRunner = func(name string, args ...string) ([]byte, error) {
		return []byte(sampleJSON), nil
	}

	d, err := GetDevice("70:F9:4A:7A:8B:CA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Name != "AirPods Max" {
		t.Errorf("expected AirPods Max, got %s", d.Name)
	}
}

func TestGetDevice_NotFound(t *testing.T) {
	orig := commandRunner
	defer func() { commandRunner = orig }()

	commandRunner = func(name string, args ...string) ([]byte, error) {
		return []byte(sampleJSON), nil
	}

	_, err := GetDevice("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent device")
	}
}

func TestParseBatteryLevel_Priority(t *testing.T) {
	// When device_batteryLevel is present, it takes priority
	props := map[string]interface{}{
		"device_batteryLevel":     "50%",
		"device_batteryLevelLeft": "100%",
	}
	got := parseBatteryLevel(props)
	if got != 50 {
		t.Errorf("expected 50 (from device_batteryLevel), got %d", got)
	}

	// When only Left is present
	props2 := map[string]interface{}{
		"device_batteryLevelLeft":  "80%",
		"device_batteryLevelRight": "90%",
		"device_batteryLevelCase":  "60%",
	}
	got2 := parseBatteryLevel(props2)
	if got2 != 80 {
		t.Errorf("expected 80 (from device_batteryLevelLeft), got %d", got2)
	}

	// When only Case is present
	props3 := map[string]interface{}{
		"device_batteryLevelCase": "45%",
	}
	got3 := parseBatteryLevel(props3)
	if got3 != 45 {
		t.Errorf("expected 45 (from device_batteryLevelCase), got %d", got3)
	}

	// When none present
	props4 := map[string]interface{}{
		"device_address": "AA:BB:CC:DD:EE:FF",
	}
	got4 := parseBatteryLevel(props4)
	if got4 != -1 {
		t.Errorf("expected -1 (unknown), got %d", got4)
	}
}

// Test with real-world-like JSON including unicode device names
const unicodeJSON = `{
  "SPBluetoothDataType" : [
    {
      "controller_properties" : {
        "controller_state" : "attrib_on"
      },
      "device_connected" : [
        {
          "胖胖的大耳机" : {
            "device_address" : "70:F9:4A:7A:8B:CA",
            "device_minorType" : "Headphones"
          }
        }
      ],
      "device_not_connected" : []
    }
  ]
}`

func TestParseDevices_UnicodeNames(t *testing.T) {
	devices, err := ParseDevices([]byte(unicodeJSON))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
	if devices[0].Name != "胖胖的大耳机" {
		t.Errorf("expected unicode name, got %s", devices[0].Name)
	}
	if !devices[0].Connected {
		t.Error("device should be connected")
	}
}

// Test with only connected devices, no disconnected section
const connectedOnlyJSON = `{
  "SPBluetoothDataType" : [
    {
      "controller_properties" : {
        "controller_state" : "attrib_on"
      },
      "device_connected" : [
        {
          "Device A" : {
            "device_address" : "AA:BB:CC:DD:EE:01",
            "device_batteryLevel" : "42%"
          }
        },
        {
          "Device B" : {
            "device_address" : "AA:BB:CC:DD:EE:02"
          }
        }
      ]
    }
  ]
}`

func TestParseDevices_ConnectedOnly(t *testing.T) {
	devices, err := ParseDevices([]byte(connectedOnlyJSON))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(devices))
	}
	for _, d := range devices {
		if !d.Connected {
			t.Errorf("device %s should be connected", d.Name)
		}
	}
}
