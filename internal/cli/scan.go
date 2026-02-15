package cli

import (
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for Bluetooth devices (alias for list)",
	Long:  "Scan for Bluetooth devices. This is an alias for 'list' since system_profiler shows all known devices.",
	RunE:  listCmd.RunE,
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
