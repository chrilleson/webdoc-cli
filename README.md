# webdoc-cli

A Go CLI for the Webdoc REST API.

## Project Structure

```
webdoc-cli/
├── cmd/
│   └── webdoc/
│       └── main.go                  # Entry point + command tree
├── internal/
│   ├── api/
│   │   └── bookingtypes.go          # Booking type models + API call
│   ├── auth/
│   │   └── auth.go                  # OAuth2 client credentials login + token validation
│   ├── config/
│   │   └── config.go                # Config persistence + URL resolution
│   └── httpclient/
│       ├── client.go                # Generic HTTP client (Get/Post/Patch)
│       └── from_config.go           # Client factory from saved config + token
├── go.mod
├── go.sum
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
