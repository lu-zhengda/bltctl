package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lu-zhengda/bltctl/internal/bluetooth"
)

type tickMsg time.Time

type deviceMsg struct {
	devices []bluetooth.Device
	err     error
}

type actionMsg struct {
	message string
	err     error
}

// keyMap defines key bindings for the TUI.
type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	Quit       key.Binding
	Connect    key.Binding
	Disconnect key.Binding
	Remove     key.Binding
	Power      key.Binding
	Reset      key.Binding
	Help       key.Binding
	Confirm    key.Binding
	Cancel     key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Up:         key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/up", "up")),
		Down:       key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/down", "down")),
		Quit:       key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Connect:    key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "connect")),
		Disconnect: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "disconnect")),
		Remove:     key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "remove")),
		Power:      key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "power toggle")),
		Reset:      key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "reset")),
		Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Confirm:    key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "confirm")),
		Cancel:     key.NewBinding(key.WithKeys("n", "esc"), key.WithHelp("n/esc", "cancel")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Connect, k.Disconnect, k.Remove, k.Power, k.Quit, k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.Connect, k.Disconnect, k.Remove},
		{k.Power, k.Reset},
		{k.Quit, k.Help},
	}
}

// Model is the Bubble Tea model for bltctl.
type Model struct {
	version    string
	keys       keyMap
	help       help.Model
	width      int
	height     int
	cursor     int
	offset     int
	devices    []bluetooth.Device
	confirming bool
	confirmMsg string
	confirmFn  func() tea.Cmd
	showHelp   bool
	err        error
	statusMsg  string
	blueutil   bool
}

// New creates a new TUI model.
func New(version string) Model {
	return Model{
		version:  version,
		keys:     newKeyMap(),
		help:     help.New(),
		blueutil: bluetooth.IsBlueUtilInstalled(),
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func fetchDevices() tea.Cmd {
	return func() tea.Msg {
		devices, err := bluetooth.ListDevices()
		return deviceMsg{devices: devices, err: err}
	}
}

func connectDevice(address, name string) tea.Cmd {
	return func() tea.Msg {
		err := bluetooth.Connect(address)
		if err != nil {
			return actionMsg{err: err}
		}
		return actionMsg{message: fmt.Sprintf("Connected to %s", name)}
	}
}

func disconnectDevice(address, name string) tea.Cmd {
	return func() tea.Msg {
		err := bluetooth.Disconnect(address)
		if err != nil {
			return actionMsg{err: err}
		}
		return actionMsg{message: fmt.Sprintf("Disconnected %s", name)}
	}
}

func removeDevice(address, name string) tea.Cmd {
	return func() tea.Msg {
		err := bluetooth.Remove(address)
		if err != nil {
			return actionMsg{err: err}
		}
		return actionMsg{message: fmt.Sprintf("Removed %s", name)}
	}
}

func togglePower(currentlyOn bool) tea.Cmd {
	return func() tea.Msg {
		var err error
		if currentlyOn {
			err = bluetooth.PowerOff()
		} else {
			err = bluetooth.PowerOn()
		}
		if err != nil {
			return actionMsg{err: err}
		}
		if currentlyOn {
			return actionMsg{message: "Bluetooth powered off"}
		}
		return actionMsg{message: "Bluetooth powered on"}
	}
}

func resetBluetooth() tea.Cmd {
	return func() tea.Msg {
		err := bluetooth.Reset()
		if err != nil {
			return actionMsg{err: err}
		}
		return actionMsg{message: "Bluetooth module reset"}
	}
}

// Init initializes the TUI.
func (m Model) Init() tea.Cmd {
	return tea.Batch(fetchDevices(), tickCmd())
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		return m, nil

	case tickMsg:
		return m, tea.Batch(fetchDevices(), tickCmd())

	case deviceMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.devices = msg.devices
		m.err = nil
		if m.cursor >= len(m.devices) && len(m.devices) > 0 {
			m.cursor = len(m.devices) - 1
		}
		return m, nil

	case actionMsg:
		if msg.err != nil {
			m.statusMsg = errorStyle.Render(fmt.Sprintf("Error: %v", msg.err))
		} else {
			m.statusMsg = statusStyle.Render(msg.message)
		}
		m.confirming = false
		return m, fetchDevices()

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If confirming an action.
	if m.confirming {
		switch {
		case key.Matches(msg, m.keys.Confirm):
			if m.confirmFn != nil {
				return m, m.confirmFn()
			}
			m.confirming = false
			return m, nil
		case key.Matches(msg, m.keys.Cancel):
			m.confirming = false
			m.statusMsg = ""
			return m, nil
		}
		return m, nil
	}

	// If showing help.
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}

	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
			if m.cursor < m.offset {
				m.offset = m.cursor
			}
		}

	case key.Matches(msg, m.keys.Down):
		max := len(m.devices) - 1
		if m.cursor < max {
			m.cursor++
			viewHeight := m.tableHeight()
			if m.cursor >= m.offset+viewHeight {
				m.offset = m.cursor - viewHeight + 1
			}
		}

	case key.Matches(msg, m.keys.Connect):
		if len(m.devices) > 0 && !m.blueutil {
			m.statusMsg = errorStyle.Render("blueutil required -- brew install blueutil")
			return m, nil
		}
		if len(m.devices) > 0 {
			d := m.devices[m.cursor]
			m.confirming = true
			m.confirmMsg = fmt.Sprintf("Connect to %s? (y/n)", d.Name)
			m.confirmFn = func() tea.Cmd {
				return connectDevice(d.Address, d.Name)
			}
		}

	case key.Matches(msg, m.keys.Disconnect):
		if len(m.devices) > 0 && !m.blueutil {
			m.statusMsg = errorStyle.Render("blueutil required -- brew install blueutil")
			return m, nil
		}
		if len(m.devices) > 0 {
			d := m.devices[m.cursor]
			m.confirming = true
			m.confirmMsg = fmt.Sprintf("Disconnect %s? (y/n)", d.Name)
			m.confirmFn = func() tea.Cmd {
				return disconnectDevice(d.Address, d.Name)
			}
		}

	case key.Matches(msg, m.keys.Remove):
		if len(m.devices) > 0 && !m.blueutil {
			m.statusMsg = errorStyle.Render("blueutil required -- brew install blueutil")
			return m, nil
		}
		if len(m.devices) > 0 {
			d := m.devices[m.cursor]
			m.confirming = true
			m.confirmMsg = fmt.Sprintf("Remove %s? This will unpair the device. (y/n)", d.Name)
			m.confirmFn = func() tea.Cmd {
				return removeDevice(d.Address, d.Name)
			}
		}

	case key.Matches(msg, m.keys.Power):
		// Determine current power state from connected devices
		hasConnected := false
		for _, d := range m.devices {
			if d.Connected {
				hasConnected = true
				break
			}
		}
		action := "on"
		if hasConnected || len(m.devices) > 0 {
			action = "off"
		}
		m.confirming = true
		m.confirmMsg = fmt.Sprintf("Turn Bluetooth %s? (y/n)", action)
		m.confirmFn = func() tea.Cmd {
			return togglePower(action == "off")
		}

	case key.Matches(msg, m.keys.Reset):
		m.confirming = true
		m.confirmMsg = "Reset Bluetooth module? (y/n)"
		m.confirmFn = func() tea.Cmd {
			return resetBluetooth()
		}

	case key.Matches(msg, m.keys.Help):
		m.showHelp = true
	}

	return m, nil
}

func (m Model) tableHeight() int {
	// Title(1) + header(1) + status(1) + help(2) + padding(2)
	overhead := 7
	h := m.height - overhead
	if h < 1 {
		h = 10
	}
	return h
}

// View renders the TUI.
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	// Title bar.
	title := fmt.Sprintf("bltctl %s", m.version)
	blueUtilStatus := ""
	if !m.blueutil {
		blueUtilStatus = dimStyle.Render(" [blueutil not installed]")
	}
	b.WriteString(titleStyle.Render(title) + blueUtilStatus)
	b.WriteString("\n")

	// Help view.
	if m.showHelp {
		b.WriteString(m.help.View(m.keys))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("Press any key to return"))
		return b.String()
	}

	// Confirm dialog.
	if m.confirming {
		b.WriteString(m.renderDeviceTable())
		b.WriteString(warnStyle.Render(m.confirmMsg))
		b.WriteString("\n")
		return b.String()
	}

	// Main content.
	b.WriteString(m.renderDeviceTable())

	// Status bar.
	if m.statusMsg != "" {
		b.WriteString(m.statusMsg)
		b.WriteString("\n")
	}

	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n")
	}

	// Short help.
	b.WriteString(m.help.View(m.keys))

	return b.String()
}

func (m Model) renderDeviceTable() string {
	var b strings.Builder

	// Header.
	header := fmt.Sprintf("%-2s %-24s %-14s %-19s %-20s",
		"", "NAME", "TYPE", "ADDRESS", "BATTERY")
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	if len(m.devices) == 0 {
		b.WriteString(dimStyle.Render("  No devices found"))
		b.WriteString("\n")
		return b.String()
	}

	viewHeight := m.tableHeight()
	end := m.offset + viewHeight
	if end > len(m.devices) {
		end = len(m.devices)
	}

	for i := m.offset; i < end; i++ {
		d := m.devices[i]

		status := "\u25cb" // disconnected
		if d.Connected {
			status = "\u25cf" // connected
		}

		deviceType := d.MinorType
		if deviceType == "" {
			deviceType = "-"
		}

		battery := "-"
		if d.BatteryLevel >= 0 {
			battery = fmt.Sprintf("%s %d%%", renderBatteryBar(d.BatteryLevel), d.BatteryLevel)
		}

		line := fmt.Sprintf("%s  %-24s %-14s %-19s %s",
			status, truncate(d.Name, 24), truncate(deviceType, 14), d.Address, battery)

		switch {
		case i == m.cursor:
			line = selectedStyle.Render(line)
		case d.Connected:
			line = connectedStyle.Render(line)
		default:
			line = disconnectedStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Footer.
	connected := 0
	for _, d := range m.devices {
		if d.Connected {
			connected++
		}
	}
	b.WriteString(dimStyle.Render(fmt.Sprintf("  %d devices (%d connected)", len(m.devices), connected)))
	b.WriteString("\n")

	return b.String()
}

// renderBatteryBar creates a colored battery bar visualization.
func renderBatteryBar(level int) string {
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
	bar = "[" + bar + "]"

	switch {
	case level > 60:
		return batteryHighStyle.Render(bar)
	case level >= 20:
		return batteryMedStyle.Render(bar)
	default:
		return batteryLowStyle.Render(bar)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "~"
}
