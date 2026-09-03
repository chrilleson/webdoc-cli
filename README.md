# webdoc-cli

A Go CLI for the Webdoc (Carasent Sweden) EMR REST API.

Covers all 41 endpoints published in the [API documentation](https://docs.webdoc.com/)
across 48 commands.

## Project Structure

```
webdoc-cli/
├── cmd/
│   └── webdoc/
│       └── main.go                  # Entry point: wires the command groups onto root
├── internal/
│   ├── api/                         # One file per resource: models + API calls
│   │   ├── common.go                # Shared models (Loose, Ref, Keyword, ...)
│   │   ├── bookings.go  visits.go  records.go  recordtemplates.go
│   │   ├── documents.go  notes.go  clinics.go  organizations.go
│   │   ├── invoices.go  paymentmethods.go  actioncodes.go  health.go
│   │   ├── users.go  patients.go  patient-types.go  bookingtypes.go
│   ├── commands/                    # One file per resource: cobra command groups
│   │   ├── commands.go              # Options{APIURL, ClinicID, JSON} + Client()/Clinic()
│   │   ├── output.go                # Emit/EmitEmpty (--json), ScopeHint
│   │   ├── flags.go                 # Date validation, query params, pagination
│   │   └── <resource>.go            # NewXxxCmd(o *Options) *cobra.Command
│   ├── auth/
│   │   └── auth.go                  # OAuth2 client credentials login + token validation
│   ├── config/
│   │   └── config.go                # Config persistence + URL/clinic resolution
│   └── httpclient/
│       ├── client.go                # Generic HTTP client (Get/Post/Patch/Put/Delete/PostMultipart)
│       └── from_config.go           # Client factory from saved config + token
├── docs/
│   ├── API-COVERAGE.md              # Every endpoint -> command, with required scope
│   ├── DESIGN.md                    # Conventions and cross-cutting decisions
│   └── TASKS.md                     # Per-endpoint implementation spec
├── go.mod
├── go.sum
└── README.md
```

## Setup

```bash
go mod tidy
go build -o webdoc ./cmd/webdoc
```

## Getting started

```bash
# Point the CLI at your environment (saved to ~/.config/webdoc/config.json)
webdoc config set-auth-url https://auth-integration.carasent.net
webdoc config set-api-url  https://api.atlan.se

# Optional: a default clinic for clinic-scoped commands
webdoc config set-clinic-id 9ce597b0-8801-4709-9bae-d732c6a10094

webdoc config show

# Authenticate (token is cached in the config file)
webdoc auth login --client-id <id> --client-secret <secret>
webdoc auth status
```

### Scopes

`auth login` defaults to `--scope self-service`, which does **not** cover every
endpoint. Eight endpoints need something else — `record-templates` and
`records create-external` need `system-admin`, for example. Commands turn a
401/403 into a hint naming the scope to re-login with. `docs/API-COVERAGE.md`
lists the required scope per endpoint.

## Global flags

| Flag | Purpose |
|---|---|
| `--url` | Override the API base URL for one command |
| `--clinic-id` | Clinic for clinic-scoped commands (overrides the config default) |
| `--json` | Print the raw API response as indented JSON instead of the text summary |

`--json` works on every command, which makes the CLI scriptable:

```bash
webdoc bookings list --from-date 2026-01-01 --json | jq '.[].id'
```

## Commands

```
auth              login | status
config            set-auth-url | set-api-url | set-clinic-id | show
health

patients          search | list | create | create-without-ssn | update
                  national-registry | set-current
patient-types     list | get <id>
users             search | list | get <id> | permissions <uuid> | list-by-clinic
clinics           list
organizations     list | create | add-patient <orgId>

bookings          list | list-by-clinic | create | update <id> | cancel <id>
booking-types     list
visits            list | create | sign <visitId>
records           update <visitId> | create-external
record-templates  list | get <id> | keywords <id>

documents         upload
document-types    list
notes             create
invoices          list | list-credit | create | update <invoiceId>
payment-methods   list
action-codes      list
```

Path parameters are positional; everything else is a flag. Run
`webdoc <group> <command> --help` for the flags of any command.

### Building record payloads

`records update` and `records create-external` take repeatable
`--keyword id=value` pairs, where the IDs are the form fields of a record
template. Look them up first — that is what `record-templates keywords` is for:

```bash
webdoc record-templates list            # find the template
webdoc record-templates keywords 42     # its field IDs, in form order

# Write into an existing visit's record
webdoc records update <visitId> --user-id <uuid> \
      --keyword 1=85 --keyword 2="No symptoms"

# Or file an external record against a patient
webdoc records create-external --patient-identity <uuid> \
      --author-identity 1010 --record-template-id 42 \
      --keyword 1=85 --diagnosis-code W00.00
```

The full booking-to-signed-visit path is:

```bash
webdoc bookings create ...              # book a slot
webdoc visits create --booking-id <id> --visit-date "2026-01-29 11:37:25"
webdoc records update <visitId> --user-id <uuid> --keyword 1=85
webdoc visits sign <visitId> --user-id <uuid>
```

`record-templates` needs the `system-admin` scope; see Scopes above.

## Config File Location

| OS      | Path                                      |
|---------|-------------------------------------------|
| Linux   | `~/.config/webdoc/config.json`            |
| macOS   | `~/.config/webdoc/config.json`            |
| Windows | `%APPDATA%\webdoc\config.json`            |

The config file is saved with `0600` permissions (owner read/write only).

## Development

There is no test suite; the quality gates are:

```bash
go build ./... && go vet ./... && gofmt -l .
```

To add a command, add the API call in `internal/api/<resource>.go` and the
cobra command in `internal/commands/<resource>.go`. Each command file exposes a
single `New<Group>Cmd(o *Options) *cobra.Command`; subcommand builders are
unexported and group-prefixed (`newBookingsListCmd`) because the package is
flat. Send all output through `Emit`/`EmitEmpty` so `--json` keeps working.
