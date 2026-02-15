package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zhengda-lu/bltctl/internal/bluetooth"
)

var connectCmd = &cobra.Command{
	Use:   "connect <device>",
	Short: "Connect to a paired device",
	Long:  "Connect to a paired Bluetooth device by name or address. Requires blueutil (brew install blueutil).",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		device, err := bluetooth.GetDevice(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("Connecting to %s (%s)...\n", device.Name, device.Address)
		if err := bluetooth.Connect(device.Address); err != nil {
			return err
		}
		fmt.Printf("Connected to %s.\n", device.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
}
