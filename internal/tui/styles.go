package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Title and header styles.
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))

	// Device row styles.
	connectedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // Green for connected
	disconnectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // Gray for disconnected
	selectedStyle     = lipgloss.NewStyle().Background(lipgloss.Color("8"))  // Highlighted row

	// Battery bar styles.
	batteryHighStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // Green >60%
	batteryMedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Yellow 20-60%
	batteryLowStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // Red <20%

	// Status and info styles.
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	labelStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	warnStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
)
