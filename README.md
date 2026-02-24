# webdoc-cli

A Go CLI for the Webdoc EMR REST API.

## Project Structure

```
webdoc-cli/
├── cmd/
│   └── webdoc/
│       └── main.go          # Entry point + command tree
├── internal/
│   └── config/
│       └── config.go        # Config persistence + URL resolution
├── go.mod
└── README.md
```

## Setup

```bash
# Install dependencies
go mod tidy

# Run directly
go run cmd/webdoc/main.go --help

# Build a binary
go build -o webdoc ./cmd/webdoc
```

## Usage

```bash
# Set your base URL once (saved to ~/.config/webdoc/config.json)
webdoc config set-url https://test.yourclinic.webdoc.com

# Check saved config
webdoc config show

# Override URL per-command (flag wins over config)
webdoc patients --url https://prod.yourclinic.webdoc.com
```

## Config File Location

| OS      | Path                                      |
|---------|-------------------------------------------|
| Linux   | `~/.config/webdoc/config.json`            |
| macOS   | `~/.config/webdoc/config.json`            |
| Windows | `%APPDATA%\webdoc\config.json`            |

The config file is saved with `0600` permissions (owner read/write only).

## Steps Completed

- [x] Step 1 — Project setup, modules, entry point
- [x] Step 2 & 3 — Cobra command tree + config system with URL resolution

## Next Steps

- [ ] Step 4 — OAuth 2.0 token fetch (`webdoc auth login`)
- [ ] Step 5 — Token caching with expiry
- [ ] Step 6 — `patients list` / `patients get`
- [ ] Step 7 — `bookings list` with date flags
- [ ] Step 8 — `documents list`
- [ ] Step 9 — Pretty table output
