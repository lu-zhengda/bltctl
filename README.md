# bltctl

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform: macOS](https://img.shields.io/badge/Platform-macOS-lightgrey.svg)](https://github.com/lu-zhengda/bltctl)
[![Homebrew](https://img.shields.io/badge/Homebrew-lu--zhengda/tap-orange.svg)](https://github.com/lu-zhengda/homebrew-tap)

macOS Bluetooth manager — browse, connect, and manage Bluetooth devices from the terminal.

## Install

```bash
brew tap lu-zhengda/tap
brew install bltctl
```

### Optional: blueutil

`bltctl` uses macOS `system_profiler` for all read operations (listing devices, battery levels, diagnostics). For **connect, disconnect, remove, and power** operations, it optionally uses [blueutil](https://github.com/toy/blueutil):

```bash
brew install blueutil
```

Without blueutil, read-only commands (`list`, `battery`, `info`, `scan`, `diagnose`) work fully. Commands that modify state will show a clear message if blueutil is missing.

## Usage

```
$ bltctl list
STATUS  NAME                  TYPE        ADDRESS            BATTERY
●       Headphones            Headphones  70:F9:4A:7A:8B:CA  -
○       AirPods               Headphones  98:DD:60:D2:4C:FF  -
○       AirPods Pro           Headphones  74:15:F5:4E:D0:50  [██████████] 100%
○       Beats Flex            Headphones  A8:91:3D:DE:91:C6  -
○       Beats Studio Buds     Headphones  F4:34:F0:96:DD:A0  [██████████] 100%

$ bltctl battery
DEVICE      BATTERY  LEVEL
AirPods Pro          [██████████] 100%
```

## Commands

| Command | Description |
|---------|-------------|
| `list` | List all paired devices with status and battery |
| `scan` | Alias for list |
| `connect <device>` | Connect to a device (name or address) |
| `disconnect <device>` | Disconnect a device |
| `battery` | Show battery levels for connected devices |
| `info <device>` | Show detailed device info |
| `remove <device>` | Unpair a device |
| `power on\|off` | Toggle Bluetooth power (requires sudo) |
| `reset` | Reset Bluetooth module (requires sudo) |
| `diagnose` | Run connection diagnostics |

## TUI

Launch `bltctl` without arguments for the interactive TUI:

- Device list with connection status and battery levels
- Color-coded: green for connected, gray for disconnected
- Battery bar visualization (green >60%, yellow 20-60%, red <20%)
- Auto-refresh every 5 seconds

| Key | Action |
|-----|--------|
| `j`/`k` | Navigate up/down |
| `c` | Connect to selected device |
| `d` | Disconnect selected device |
| `r` | Remove (unpair) selected device |
| `p` | Toggle Bluetooth power |
| `R` | Reset Bluetooth module |
| `q` | Quit |

## Claude Code

Available as a skill in the [macos-toolkit](https://github.com/lu-zhengda/macos-toolkit) Claude Code plugin.

## License

[MIT](LICENSE)
