package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/lu-zhengda/bltctl/internal/tui"
)

var (
	// version is set via ldflags at build time.
	version = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "bltctl",
	Short: "macOS Bluetooth manager",
	Long: `bltctl is a macOS Bluetooth manager — browse, connect, and manage
Bluetooth devices with a live-updating TUI or handy CLI subcommands.
Launch without subcommands for interactive TUI mode.`,
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		if shell, _ := cmd.Flags().GetString("generate-completion"); shell != "" {
			switch shell {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			default:
				return fmt.Errorf("unsupported shell: %s (use bash, zsh, or fish)", shell)
			}
		}
		p := tea.NewProgram(tui.New(version), tea.WithAltScreen())
		_, err := p.Run()
		return err
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("bltctl %s\n", version))
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.Flags().String("generate-completion", "", "Generate shell completion (bash, zsh, fish)")
	rootCmd.Flags().MarkHidden("generate-completion")
}
