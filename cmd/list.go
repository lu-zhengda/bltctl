package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/zhengda-lu/bltctl/internal/bluetooth"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List paired Bluetooth devices",
	Long:  "List all paired Bluetooth devices with name, type, connection status, and battery level.",
	RunE: func(cmd *cobra.Command, args []string) error {
		devices, err := bluetooth.ListDevices()
		if err != nil {
			return err
		}

		if len(devices) == 0 {
			fmt.Println("No paired devices found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "STATUS\tNAME\tTYPE\tADDRESS\tBATTERY")

		for _, d := range devices {
			status := "\u25cb"
			if d.Connected {
				status = "\u25cf"
			}

			deviceType := d.MinorType
			if deviceType == "" {
				deviceType = "-"
			}

			battery := "-"
			if d.BatteryLevel >= 0 {
				battery = fmt.Sprintf("%s %d%%", batteryBar(d.BatteryLevel), d.BatteryLevel)
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				status, d.Name, deviceType, d.Address, battery)
		}

		return w.Flush()
	},
}

// batteryBar returns a visual bar representation of battery level.
func batteryBar(level int) string {
	const barLen = 10
	filled := level * barLen / 100
	bar := ""
	for i := 0; i < barLen; i++ {
		if i < filled {
			bar += "\u2588"
		} else {
			bar += "\u2591"
		}
	}
	return "[" + bar + "]"
}

func init() {
	rootCmd.AddCommand(listCmd)
}
