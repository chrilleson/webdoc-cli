# webdoc-cli design

Read this before writing any command. The shared seam described here is **already
implemented and building** — do not re-plan it, just follow it.

## 1. Command tree

Consistent shape: **plural resource noun** as the group, **verb** as the subcommand.
Path parameters are positional args; everything else is a flag.

```
webdoc
  auth              login | status
  config            set-auth-url | set-api-url | set-clinic-id | show
  health

  action-codes      list
  clinics           list

  bookings          list | list-by-clinic | create | update <bookingId> | cancel <bookingId>
  booking-types     list

  visits            list | create | sign <visitId>
  records           update <visitId> | create-external
  record-templates  list | get <templateId> | keywords <templateId>

  documents         upload
  document-types    list
  notes             create

  patients          search | list | create | create-without-ssn | update <personalNumber>
                    | national-registry <personalNumber> | set-current <personalNumber>
  patient-types     list | get <id>

  users             search | list | get <userId> | permissions <uuid> | list-by-clinic

  organizations     list | create | add-patient <organizationId>
  invoices          list | list-credit | create | update <invoiceId>
  payment-methods   list
```

Naming rules:

- `list` = collection GET. When the same resource has both a global and a
  clinic-scoped collection endpoint, the global one is `list` and the scoped one is
  `list-by-clinic` (`bookings`, `users`). Never overload `list` with hidden routing
  based on whether `--clinic-id` happens to be set.
- `get <id>` = single-resource GET by path parameter.
- `create` / `update` / `cancel` = POST / PATCH / DELETE.
- `search` is reserved for the two pre-existing commands (`patients search`,
  `users search`); do not add new `search` verbs.
- Group names are kebab-case plurals of the API noun (`record-templates`,
  `payment-methods`, `document-types`).

### Record templates are read-only — on purpose

The collection exposes exactly two record-template endpoints, both GET, with no
documented query parameters. **There is no create/update/delete for record
templates anywhere in the API.** Do not invent them. "Managing" record templates
means list / get / inspect, nothing more.

They are nonetheless **P0**, because they are the discovery dependency for the whole
P0 write path: `POST /v2/visits`, `PATCH /v1/visits/:visitId/records` and
`POST /v1/externalRecords` all take `recordTemplateId` and `keywords[{id,value}]`,
and the only way to learn the valid keyword IDs for a template is
`GET /v1/recordTemplates/:templateId`. Hence the third subcommand,
`record-templates keywords <templateId>`, which prints the flat keyword lookup you
need to hand-build a records payload.

`templateType` and keyword `type` are undocumented integer enums. Print the raw
integers; do not invent label mappings.

## 2. The seam: `internal/commands`

`cmd/webdoc/main.go` used to hold the entire tree in one 380-line `main()`. It is now
~35 lines and does nothing but declare persistent flags and list constructors.

**Layout — one file per resource group, and each file is owned by exactly one person.**

```
cmd/webdoc/main.go              root command + persistent flags + constructor list
internal/commands/commands.go   Options, Options.Client(), Options.Clinic()      [frozen]
internal/commands/output.go     Emit, EmitEmpty, PrintJSON, ScopeHint            [frozen]
internal/commands/flags.go      date/time validation, query + pagination helpers [frozen]
internal/commands/<resource>.go one per group
internal/api/common.go          shared nested response types                     [frozen]
internal/api/<resource>.go      models + request funcs, one per resource
```

### The pattern every command file must follow

One exported constructor per group, named `New<Group>Cmd`, taking `*Options`:

```go
package commands

func NewBookingsCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bookings",
		Short: "Manage bookings",
	}
	cmd.AddCommand(
		newBookingsListCmd(o),
		newBookingsCreateCmd(o),
	)
	return cmd
}

func newBookingsCreateCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a booking",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			userID, _ := cmd.Flags().GetString("user-id")
			// ... read the rest of the flags

			booking, err := api.CreateBooking(client, req)
			if err != nil {
				return err
			}

			return Emit(o, booking, func() {
				fmt.Printf("Booking created: %s\n", booking.ID)
			})
		},
	}
	cmd.Flags().String("user-id", "", "Caregiver user ID (required)")
	cmd.MarkFlagRequired("user-id")
	return cmd
}
```

Rules, so the two workstreams compose:

1. **Exactly one exported symbol per file**: `New<Group>Cmd(o *Options) *cobra.Command`.
   Every subcommand builder is unexported and prefixed with the group name
   (`newBookingsListCmd`, not `newListCmd`) — the package is flat, so unprefixed
   names will collide.
2. **Read flag values inside `RunE`**, via `cmd.Flags().GetX(...)`. Do not bind to
   local variables captured at construction time except where a `StringVar` is
   already used (auth). This keeps every builder independent.
3. **Never edit `main.go`.** Add your group's constructor line yourself, and only
   that line — it is a one-line, trivially-mergeable addition to an alphabetical
   list. Everything else in `main.go` stays untouched.
4. **Never edit another owner's `<resource>.go`.** Ownership is in `TASKS.md`.
5. **Never edit the four `[frozen]` files.** If you need a new shared helper, say so
   and the architect adds it; do not add package-level helpers in a resource file.

### `Options`

```go
type Options struct {
	APIURL   string // --url
	ClinicID string // --clinic-id
	JSON     bool   // --json
}

func (o *Options) Client() (*httpclient.Client, error) // authed client, honours --url
func (o *Options) Clinic() (string, error)             // --clinic-id, else config clinic_id, else error
```

## 3. `internal/httpclient` additions (done)

```go
func Get[T any](c *Client, path string, queryParams url.Values) (T, error)
func Post[T any](c *Client, path string, body any) (T, error)
func Patch[T any](c *Client, path string, body any) (T, error)
func Put[T any](c *Client, path string, body any) (T, error)      // NEW
func Delete[T any](c *Client, path string, body any) (T, error)   // NEW
func PostMultipart[T any](c *Client, path, contentType string, body io.Reader) (T, error)
```

- `Delete` takes a body because `DELETE /v1/bookings/:bookingId` **requires** one
  (`{"userId": "..."}`). Pass `nil` when an endpoint takes no body — `sendJSON` now
  omits the body and the `Content-Type` header entirely rather than sending `null`.
- For endpoints with no response body (204 / empty), instantiate with
  `httpclient.Delete[struct{}]` or `[json.RawMessage]` and ignore the value.

## 4. Output conventions

Every command returns through **one** of these, never a bare `fmt.Printf` chain:

```go
// single object
return Emit(o, obj, func() { fmt.Printf(...) })

// collection
return EmitEmpty(o, items, len(items) == 0, "No bookings found.", func() {
	for _, it := range items { fmt.Printf(...) }
})
```

`--json` is a global persistent flag and prints the decoded response as indented
JSON, which makes the CLI scriptable (`webdoc bookings list --json | jq`). With
`--json`, an empty collection still prints `[]` rather than the prose message.

Human formatting follows the existing house style — plain `fmt.Printf`, no
tabwriter, no colour:

- Lists: one line per record, id first, fixed-width columns via `%-Ns`.
- Detail views: a header line, then two-space-indented `Label: value` lines.
- Nested children: `  └─ ` prefix (as in `users search`).
- Omit optional fields that are empty rather than printing `Email: `.
- Creates print `Xxx created: <id>` then a short summary.
- Updates/deletes with no response body print a single confirmation line, e.g.
  `Booking 123 cancelled`.

Errors: return them, don't print them. `httpclient` already wraps non-2xx into
`*APIError`, and the root command has `SilenceUsage: true` so an API failure does not
dump the help text. For any endpoint whose documented scope is not `self-service`,
wrap the error with `ScopeHint`:

```go
if err != nil {
	return ScopeHint(err, "system-admin")
}
```

which turns a 401/403 into a message telling the user to re-login with that scope.
The scope for every endpoint is in `API-COVERAGE.md`.

## 5. Clinic-scoped endpoints

Nine endpoints are under `/v1/clinics/:clinicId/...`. They resolve the clinic via
`o.Clinic()`: the `--clinic-id` persistent flag wins, otherwise the persisted
`clinic_id` from `webdoc config set-clinic-id <id>`, otherwise a clear error. Never
make `clinicId` a positional arg and never add a per-command `--clinic-id` flag —
the persistent one already exists.

`clinicId` is *also* an optional **query filter** on `GET /v1/bookings`. That is a
different thing: keep it as a normal `--clinic-filter-id`-style flag on that command,
or reuse `o.ClinicID` directly without calling `o.Clinic()` (so it stays optional).

## 6. Cross-cutting conventions

**Dates and times.** Webdoc uses three wire formats. Validate before sending so a
typo fails locally instead of as an opaque 400:

| Layout | Where |
|---|---|
| `2006-01-02` | `date`, `fromDate`, `toDate`, `birthDate` |
| `15:04` | `startTime`, `endTime` |
| `2006-01-02 15:04:05` | `visitDate` (`POST /v2/visits`), `createdAt` (notes, documents) |

Use `ValidateDate` / `ValidateTime` / `ValidateDateTime` from `flags.go`. They pass
`""` through unchanged, so they work on optional filter flags too.

**Pagination.** `GET /v1/clinics/:clinicId/bookings` documents `limit` + `offset`;
`GET /v1/bookings` documents `offset`. Register them with `AddPagination(cmd)` and
copy them with `SetPagination(cmd, q)`. Do not paginate transparently in a loop — one
command invocation is one request.

**Query building.** Use `SetStr` / `SetInt` (skip empty/zero) and `SetBool` /
`SetBool01` (only emit when the user actually passed the flag). Several Webdoc filters
are `1`/`0` rather than `true`/`false` (`arrived`, `isBooked`, `isExported`,
`isSigned`, `arrivalStatus`) — that is what `SetBool01` is for. A zero-valued flag
must never leak into the query string as a filter the user did not ask for.

**Optional fields: pointer vs value.**

- *Request* structs: value types with `,omitempty` for strings and slices. Use a
  **pointer** (`*int`, `*bool`, `*float64`) only where the API distinguishes "absent"
  from the zero value — e.g. `allowOverlap: false` and `fee: 0` in
  `POST /v1/bookings` are meaningful values, so those are `*bool` / `*float64`, set
  only when `cmd.Flags().Changed(name)`.
- *Response* structs: value types throughout, except objects that can legitimately be
  `null` (`recordType`, `actionCode`, `organization`), which are pointers so you can
  nil-check before dereferencing.

**Loose numeric JSON.** Webdoc is inconsistent about string-vs-number for the same
logical field across endpoints (`price` is a number on bookings and a string on
invoices; `VATPercentage` is both; `maxLength` is the string `"64"`). `api.Loose` in
`internal/api/common.go` decodes either and prints what the API sent. Use it for any
field the examples show as both, otherwise `json.Unmarshal` will fail at runtime on a
perfectly valid response.

**Shared nested types.** `internal/api/common.go` holds the types that appear inside
more than one resource: `Loose`, `Ref`, `IDRef`, `Caregiver`, `Payment`, `FreeCard`,
`PatientListing`, `EmbeddedPatient`, `Keyword`. Reuse them. Declaring your own
`Caregiver` or `Payment` in a resource file will break the build for the other
engineer, since `api` is one flat package.

**Known collision to avoid:** `internal/api/bookingtypes.go` already declares
`type ActionCode struct { ID int; Name string }`, which is *not* the shape of
`GET /v1/actionCodes`. Name the new one `ActionCodeDetail`.

## 7. Quality gates

There is no test suite and no linter config. Before handing work over:

```sh
go build ./...
go vet ./...
gofmt -l .            # must print nothing
go run ./cmd/webdoc --help
go run ./cmd/webdoc <your-group> --help
```
