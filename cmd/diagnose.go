package cmd

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/zhengda-lu/bltctl/internal/bluetooth"
)

var diagnoseCmd = &cobra.Command{
	Use:   "diagnose",
	Short: "Run Bluetooth diagnostics",
	Long:  "Run a comprehensive Bluetooth diagnostic check including power state, controller info, devices, and recent errors.",
	RunE: func(cmd *cobra.Command, args []string) error {
		report, err := bluetooth.Diagnose()
		if err != nil {
			return err
		}

		// Power state
		fmt.Printf("Power State: %s\n\n", report.PowerState)

		// Controller info
		if len(report.ControllerInfo) > 0 {
			fmt.Println("Controller Info:")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

			// Sort keys for stable output
			keys := make([]string, 0, len(report.ControllerInfo))
			for k := range report.ControllerInfo {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				fmt.Fprintf(w, "  %s\t%s\n", k, report.ControllerInfo[k])
			}
			w.Flush()
			fmt.Println()
		}

		// Connected devices
		if len(report.ConnectedDevices) > 0 {
			fmt.Println("Connected Devices:")
			for _, d := range report.ConnectedDevices {
				battery := ""
				if d.BatteryLevel >= 0 {
					battery = fmt.Sprintf(" [%d%%]", d.BatteryLevel)
				}
				fmt.Printf("  %s (%s)%s\n", d.Name, d.Address, battery)
			}
			fmt.Println()
		} else {
			fmt.Println("Connected Devices: none")
		fmt.Println()
		}

		// Recent errors
		if len(report.RecentErrors) > 0 {
			fmt.Println("Recent Errors (last 5 min):")
			for _, e := range report.RecentErrors {
				fmt.Printf("  %s\n", e)
			}
		} else {
			fmt.Println("Recent Errors: none")
		}

		// BlueUtil availability
		if bluetooth.IsBlueUtilInstalled() {
			fmt.Println("\nblueutil: installed")
		} else {
			fmt.Println("\nblueutil: not installed (brew install blueutil for connect/disconnect)")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(diagnoseCmd)
}
