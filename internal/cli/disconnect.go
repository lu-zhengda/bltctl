package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/bltctl/internal/bluetooth"
)

var disconnectCmd = &cobra.Command{
	Use:   "disconnect <device>",
	Short: "Disconnect a device",
	Long:  "Disconnect a Bluetooth device by name or address. Requires blueutil (brew install blueutil).",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		device, err := bluetooth.GetDevice(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("Disconnecting %s (%s)...\n", device.Name, device.Address)
		if err := bluetooth.Disconnect(device.Address); err != nil {
			return err
		}
		fmt.Printf("Disconnected %s.\n", device.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(disconnectCmd)
}
