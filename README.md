# Panopticon

Real-time Apple Silicon terminal dashboard. Monitor CPU, GPU, memory, disk I/O, and thermal state — no sudo required.

Built with [bubbletea](https://github.com/charmbracelet/bubbletea) and [lipgloss](https://github.com/charmbracelet/lipgloss), using cgo to tap directly into Mach and IOKit APIs.

## Why

`top` doesn't show GPU. `powermetrics` needs sudo. Activity Monitor is a GUI. None of them show unified memory pressure, GPU compute vs graphics breakdown, or per-process GPU time in one place.

Panopticon does — 5Hz, no sudo, one terminal. Optionally integrates with [Aurelia](https://github.com/benaskins/aurelia) for service management alongside hardware metrics.

## Features

- **CPU heatmap** — per-core utilization with P-core/E-core clustering
- **GPU monitoring** — tiler, renderer, and compute workload tracking with per-process breakdown
- **Memory** — real-time usage with top process breakdown
- **Disk I/O** — read/write rates in MB/s
- **Thermal state** — nominal through critical, color-coded
- **Aurelia integration** — optional service management over Unix socket (start/stop/restart, log tailing)

## Requirements

- macOS (Apple Silicon)
- Go 1.25+
- cgo enabled (default)

Builds on other platforms with stub implementations that return zero values.

## Install

```bash
go build -o pan ./cmd/pan/
cp pan ~/bin/  # or wherever you keep local binaries
```

## Usage

```bash
pan
```

### Keybindings

| Key | Action |
|-----|--------|
| `tab` | Toggle focus (hardware / services) |
| `j` / `k` | Navigate services |
| `enter` | Toggle service logs |
| `s` / `x` / `r` | Start / stop / restart service |
| `?` | Help |
| `q` | Quit |

## Architecture

```
cmd/pan/           CLI entry point
internal/
  hw/              cgo wrappers for Mach, IOKit, Foundation
  aurelia/         HTTP-over-Unix-socket client for Aurelia daemon
  ui/              bubbletea TUI — panels, layout, styling
```

Hardware data is polled at 5Hz with EMA smoothing for stable display. Aurelia service state polls at 1s with graceful degradation when the daemon isn't running.

### Data Sources

| Data | API |
|------|-----|
| CPU topology | `sysctlbyname()` |
| CPU per-core | `host_processor_info()` |
| GPU utilization | IOKit `AGXAcceleratorG15X` |
| GPU clients | IOKit `AGXDeviceUserClient` |
| Memory | `host_statistics64()` |
| Memory processes | `ps` |
| Disk I/O | IOKit `IOBlockStorageDriver` |
| Thermal | `NSProcessInfo.thermalState` |

## Development

```bash
just build          # build binary
just test           # run tests
just check          # format + vet
```
