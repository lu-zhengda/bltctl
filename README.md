# bltctl

macOS Bluetooth manager -- browse, connect, and manage Bluetooth devices with a live-updating TUI or handy CLI subcommands.

## Install

```bash
# Homebrew (after first release)
brew install lu-zhengda/tap/bltctl

# From source
go install github.com/zhengda-lu/bltctl@latest

# Build locally
git clone https://github.com/zhengda-lu/bltctl.git
cd bltctl && go build -o bltctl .
```

### Optional: blueutil

`bltctl` uses macOS `system_profiler` for all read operations (listing devices, battery levels, diagnostics). For **connect, disconnect, remove, and power** operations, it optionally uses [blueutil](https://github.com/toy/blueutil):

```bash
brew install blueutil
```

Without blueutil, read-only commands (`list`, `battery`, `info`, `scan`, `diagnose`) work fully. Commands that modify state (`connect`, `disconnect`, `remove`) will show a clear message if blueutil is missing.

## Quick Start

```bash
# Launch interactive TUI
bltctl

# List all paired devices
bltctl list

# Show battery levels
bltctl battery

# Connect to a device by name
bltctl connect "AirPods Pro"

# Disconnect a device
bltctl disconnect "AirPods Pro"

# Run diagnostics
bltctl diagnose
```

## Commands

| Command | Description |
|---|---|
| `bltctl` | Launch interactive TUI |
| `bltctl list` | List all paired devices |
| `bltctl scan` | Alias for list |
| `bltctl connect <device>` | Connect to a device (name or address) |
| `bltctl disconnect <device>` | Disconnect a device |
| `bltctl battery` | Show battery levels for connected devices |
| `bltctl info <device>` | Show detailed device info |
| `bltctl remove <device>` | Unpair a device |
| `bltctl power on\|off` | Toggle Bluetooth power (requires sudo) |
| `bltctl reset` | Reset Bluetooth module (requires sudo) |
| `bltctl diagnose` | Run connection diagnostics |

## TUI

Launch without subcommands for the interactive TUI:

```
bltctl
```

**Features:**
- Device list with connection status and battery levels
- Color-coded: green for connected, gray for disconnected
- Battery bar visualization (green >60%, yellow 20-60%, red <20%)
- Auto-refresh every 5 seconds
- Status bar showing blueutil availability

**Keybindings:**

| Key | Action |
|---|---|
| `j`/`k` | Navigate up/down |
| `c` | Connect to selected device |
| `d` | Disconnect selected device |
| `r` | Remove (unpair) selected device |
| `p` | Toggle Bluetooth power |
| `R` | Reset Bluetooth module |
| `?` | Show help |
| `q` | Quit |

## License

MIT
