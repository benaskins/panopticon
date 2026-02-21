# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Build & Test Commands

```bash
# Build
go build -o pan ./cmd/pan/

# Test
go test ./...                                          # all unit tests
go test ./internal/hw/                                 # single package
go test -v ./...                                       # verbose output

# Format & vet
go fmt ./...
go vet ./...
```

## Architecture

Panopticon (`pan`) is a **real-time Apple Silicon terminal dashboard** built with bubbletea/lipgloss. It displays CPU, GPU, memory, disk I/O, and thermal data via cgo (Mach APIs, IOKit), with optional Aurelia service integration over Unix socket.

### Layers

1. **Hardware** (`internal/hw`) — cgo wrappers for macOS system APIs (Mach, IOKit, Foundation). Darwin-only with stubs for other platforms.
2. **Aurelia client** (`internal/aurelia`) — HTTP-over-Unix-socket client to poll Aurelia daemon for service state and logs. Graceful degradation when daemon not running.
3. **UI** (`internal/ui`) — bubbletea TUI with lipgloss styling. Panel-based layout: memory, CPU, GPU, services, logs, status bar.
4. **CLI** (`cmd/pan`) — entry point, wires everything together.

### Key data sources (all no-sudo, darwin-only)

| Data | API |
|------|-----|
| CPU topology | `sysctlbyname()` |
| CPU per-core | `host_processor_info()` (Mach) |
| GPU utilization | IOKit `AGXAcceleratorG15X` properties |
| GPU clients | IOKit `AGXDeviceUserClient` entries |
| Memory | `host_statistics64()` (Mach) |
| Disk I/O | IOKit `IOBlockStorageDriver` statistics |
| Thermal | `NSProcessInfo.thermalState` |

### Platform Constraints

- cgo required for all hardware packages — darwin-only build tags
- Non-darwin platforms get stub implementations returning zero values

## Commit Conventions

Conventional commits: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `infra:`, `config:`
