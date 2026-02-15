package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/bltctl/internal/bluetooth"
)

var removeCmd = &cobra.Command{
	Use:   "remove <device>",
	Short: "Unpair a device",
	Long:  "Unpair a Bluetooth device by name or address. Requires blueutil (brew install blueutil).",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		device, err := bluetooth.GetDevice(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("Removing %s (%s)...\n", device.Name, device.Address)
		if err := bluetooth.Remove(device.Address); err != nil {
			return err
		}
		fmt.Printf("Removed %s.\n", device.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
