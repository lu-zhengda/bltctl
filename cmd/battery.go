package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/zhengda-lu/bltctl/internal/bluetooth"
)

var batteryCmd = &cobra.Command{
	Use:   "battery",
	Short: "Show battery levels for connected devices",
	Long:  "Show battery levels for all connected Bluetooth devices.",
	RunE: func(cmd *cobra.Command, args []string) error {
		devices, err := bluetooth.ListConnected()
		if err != nil {
			return err
		}

		if len(devices) == 0 {
			fmt.Println("No connected devices.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "DEVICE\tBATTERY\tLEVEL")

		for _, d := range devices {
			level := "unknown"
			bar := ""
			if d.BatteryLevel >= 0 {
				level = fmt.Sprintf("%d%%", d.BatteryLevel)
				bar = batteryBar(d.BatteryLevel)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", d.Name, bar, level)
		}

		return w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(batteryCmd)
}
