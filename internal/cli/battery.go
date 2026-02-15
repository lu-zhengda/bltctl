package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/bltctl/internal/bluetooth"
)

// BatteryAlert represents a low-battery alert for JSON output.
type BatteryAlert struct {
	Device       string `json:"device"`
	BatteryLevel int    `json:"battery_level"`
	Threshold    int    `json:"threshold"`
	Alert        string `json:"alert"`
}

// BatteryWatchOutput represents a single watch poll result for JSON output.
type BatteryWatchOutput struct {
	Timestamp string              `json:"timestamp"`
	Devices   []bluetooth.Device  `json:"devices"`
	Alerts    []BatteryAlert      `json:"alerts,omitempty"`
}

var batteryCmd = &cobra.Command{
	Use:   "battery",
	Short: "Show battery levels for connected devices",
	Long:  "Show battery levels for all connected Bluetooth devices.",
	RunE:  runBattery,
}

func runBattery(cmd *cobra.Command, args []string) error {
	watch, _ := cmd.Flags().GetBool("watch")
	if !watch {
		return showBattery()
	}

	interval, _ := cmd.Flags().GetInt("interval")
	low, _ := cmd.Flags().GetInt("low")
	return watchBattery(interval, low)
}

// showBattery displays battery levels once and returns.
func showBattery() error {
	devices, err := bluetooth.ListConnected()
	if err != nil {
		return err
	}

	if jsonFlag {
		return printJSON(devices)
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
}

// watchBattery polls battery levels at the given interval and alerts on low battery.
// Returns a non-nil error (with exit code 1) if any device drops below the threshold.
func watchBattery(intervalSec, threshold int) error {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sig)

	ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
	defer ticker.Stop()

	// Run immediately on first tick
	if err := pollBattery(threshold); err != nil {
		return err
	}

	for {
		select {
		case <-sig:
			return nil
		case <-ticker.C:
			if err := pollBattery(threshold); err != nil {
				return err
			}
		}
	}
}

// pollBattery checks battery levels once and returns an error if any device is below threshold.
func pollBattery(threshold int) error {
	devices, err := bluetooth.ListConnected()
	if err != nil {
		return err
	}

	var alerts []BatteryAlert
	for _, d := range devices {
		if d.BatteryLevel >= 0 && d.BatteryLevel < threshold {
			alerts = append(alerts, BatteryAlert{
				Device:       d.Name,
				BatteryLevel: d.BatteryLevel,
				Threshold:    threshold,
				Alert:        "low battery",
			})
		}
	}

	if jsonFlag {
		output := BatteryWatchOutput{
			Timestamp: time.Now().Format(time.RFC3339),
			Devices:   devices,
			Alerts:    alerts,
		}
		if err := printJSON(output); err != nil {
			return err
		}
	} else {
		now := time.Now().Format("15:04:05")
		fmt.Printf("[%s] Polling battery levels...\n", now)

		if len(devices) == 0 {
			fmt.Println("  No connected devices.")
		} else {
			for _, d := range devices {
				level := "unknown"
				if d.BatteryLevel >= 0 {
					level = fmt.Sprintf("%d%%", d.BatteryLevel)
				}
				fmt.Printf("  %s: %s %s\n", d.Name, batteryBar(d.BatteryLevel), level)
			}
		}

		for _, a := range alerts {
			fmt.Printf("  WARNING: %s battery at %d%% (below %d%% threshold)\n",
				a.Device, a.BatteryLevel, a.Threshold)
		}
	}

	if len(alerts) > 0 {
		return fmt.Errorf("low battery detected on %d device(s)", len(alerts))
	}

	return nil
}

func init() {
	batteryCmd.Flags().Bool("watch", false, "Continuously monitor battery levels")
	batteryCmd.Flags().Int("interval", 30, "Polling interval in seconds (used with --watch)")
	batteryCmd.Flags().Int("low", 20, "Low battery threshold percentage (used with --watch)")
	rootCmd.AddCommand(batteryCmd)
}
