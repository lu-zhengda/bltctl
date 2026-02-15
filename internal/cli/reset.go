package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/bltctl/internal/bluetooth"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset Bluetooth module",
	Long:  "Reset the Bluetooth module by killing bluetoothd (auto-restarts). Requires sudo.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Resetting Bluetooth module...")
		if err := bluetooth.Reset(); err != nil {
			return err
		}
		fmt.Println("Bluetooth module reset. It may take a moment to reconnect devices.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)
}
