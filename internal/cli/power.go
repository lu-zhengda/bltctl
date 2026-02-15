package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/bltctl/internal/bluetooth"
)

var powerCmd = &cobra.Command{
	Use:   "power <on|off>",
	Short: "Toggle Bluetooth power",
	Long:  "Turn Bluetooth on or off. Requires sudo.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "on":
			fmt.Println("Turning Bluetooth on...")
			if err := bluetooth.PowerOn(); err != nil {
				return err
			}
			fmt.Println("Bluetooth is now on.")
		case "off":
			fmt.Println("Turning Bluetooth off...")
			if err := bluetooth.PowerOff(); err != nil {
				return err
			}
			fmt.Println("Bluetooth is now off.")
		default:
			return fmt.Errorf("invalid argument: %s (use 'on' or 'off')", args[0])
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(powerCmd)
}
