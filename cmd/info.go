package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/bltctl/internal/bluetooth"
)

var infoCmd = &cobra.Command{
	Use:   "info <device>",
	Short: "Show detailed device info",
	Long:  "Show detailed information for a specific Bluetooth device by name or address.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		device, err := bluetooth.GetDevice(args[0])
		if err != nil {
			return err
		}

		status := "Disconnected"
		if device.Connected {
			status = "Connected"
		}

		fmt.Printf("Name:       %s\n", device.Name)
		fmt.Printf("Address:    %s\n", device.Address)
		fmt.Printf("Type:       %s\n", valueOrDash(device.MinorType))
		fmt.Printf("Status:     %s\n", status)

		if device.BatteryLevel >= 0 {
			fmt.Printf("Battery:    %s %d%%\n", batteryBar(device.BatteryLevel), device.BatteryLevel)
		} else {
			fmt.Printf("Battery:    -\n")
		}

		if device.RSSI != 0 {
			fmt.Printf("RSSI:       %d dBm\n", device.RSSI)
		} else {
			fmt.Printf("RSSI:       -\n")
		}

		return nil
	},
}

func valueOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
