package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/bltctl/internal/bluetooth"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show Bluetooth connect/disconnect events",
	Long:  "Show recent Bluetooth connect/disconnect events from the system log.",
	RunE: func(cmd *cobra.Command, args []string) error {
		duration, _ := cmd.Flags().GetString("last")

		events, err := bluetooth.FetchHistory(duration)
		if err != nil {
			return err
		}

		if jsonFlag {
			return printJSON(events)
		}

		if len(events) == 0 {
			fmt.Println("No Bluetooth events found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TIMESTAMP\tEVENT\tDEVICE")

		for _, ev := range events {
			ts := ""
			if !ev.Timestamp.IsZero() {
				ts = ev.Timestamp.Format("2006-01-02 15:04:05")
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", ts, ev.EventType, ev.Device)
		}

		return w.Flush()
	},
}

func init() {
	historyCmd.Flags().String("last", "24h", "Time range to search (e.g. 1h, 30m, 7d)")
	rootCmd.AddCommand(historyCmd)
}
