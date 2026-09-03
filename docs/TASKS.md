# Workstreams

Read `DESIGN.md` first. The shared seam is already implemented, builds clean, and the
existing commands still work — you only add new files.

**File ownership.** The two workstreams touch disjoint files. The only file both of
you append to is `cmd/webdoc/main.go`, and only to add one line to the alphabetical
`rootCmd.AddCommand(...)` list — a one-line, trivially-mergeable change.

| Owner | `internal/api/` | `internal/commands/` |
|---|---|---|
| architect (frozen) | `common.go` | `commands.go`, `output.go`, `flags.go`, `auth.go`, `config.go` |
| engineer 1 | `bookings.go`, `visits.go`, `records.go`, `recordtemplates.go`, `documents.go`, `notes.go` | `bookings.go`, `visits.go`, `records.go`, `recordtemplates.go`, `documents.go`, `documenttypes.go`, `notes.go` |
| engineer 2 | `clinics.go`, `users.go`, `patients.go`, `patients_types.go`, `organizations.go`, `invoices.go`, `paymentmethods.go`, `actioncodes.go`, `health.go` | `clinics.go`, `users.go`, `patients.go`, `organizations.go`, `invoices.go`, `paymentmethods.go`, `actioncodes.go`, `health.go` |
| nobody (already done) | `bookingtypes.go`, `patient-types.go` | `bookingtypes.go`, `patienttypes.go` |

Engineer 2 is the only person who edits `internal/api/users.go`,
`internal/api/patients.go`, `internal/api/patients_types.go`,
`internal/commands/users.go` and `internal/commands/patients.go` — all of which
already exist and are extended, not replaced.

Types shared across both workstreams live in `internal/api/common.go`
(`Loose`, `Ref`, `IDRef`, `Caregiver`, `Payment`, `FreeCard`, `PatientListing`,
`EmbeddedPatient`, `Keyword`). Reuse them; `api` is one flat package and a duplicate
`type Caregiver` breaks the other engineer's build.

---

# Engineer 1 — the P0 write path

Do the tasks in order. **E1-1 first**: you cannot construct a valid visit or record
payload without reading a template's keyword IDs, so record templates unblock
everything after it.

All of E1-1 and E1-9 need `--scope system-admin`; E1-2/E1-6 need `bookings:write`,
E1-4 needs `patient:write`, E1-5 needs `documents:write`, E1-3's sign needs
`record-signatures:write`. Wrap those errors in `ScopeHint(err, "<scope>")`.

## E1-1 · record-templates (P0)

`GET /v1/recordTemplates` · `GET /v1/recordTemplates/:templateId` · scope `system-admin`
Files: `internal/api/recordtemplates.go`, `internal/commands/recordtemplates.go`

Neither endpoint has query parameters. There is **no** create/update/delete for record
templates in the API — do not add any.

```go
type RecordTemplate struct {
	ID               int                 `json:"id"`
	Title            string              `json:"title"`
	TemplateType     int                 `json:"templateType"`
	PreviousVersions []json.RawMessage   `json:"previousVersions"` // always [] in the examples
	Keywords         []TemplateKeyword   `json:"keywords"`
}

type TemplateKeyword struct {
	ID             int           `json:"id"`
	Title          string        `json:"title"`
	Type           int           `json:"type"`
	MaxLength      string        `json:"maxLength"`    // STRING in the JSON: "64"
	DefaultValue   string        `json:"defaultValue"` // STRING in the JSON: "70"
	ParentID       int           `json:"parentId"`
	WarnIfEmpty    bool          `json:"warnIfEmpty"`
	Information    string        `json:"information"`
	RepetitionTime string        `json:"repetitionTime"`
	WarningTime    string        `json:"warningTime"`
	Visual         KeywordVisual `json:"visual"`
}

type KeywordVisual struct {
	Order int `json:"order"`
	SizeX int `json:"sizeX"`
	SizeY int `json:"sizeY"`
}

func ListRecordTemplates(c *httpclient.Client) ([]RecordTemplate, error)
func GetRecordTemplate(c *httpclient.Client, templateID string) (RecordTemplate, error)
```

**Three traps — get these right:**

1. **`GET /v1/recordTemplates/:templateId` returns a JSON *array*, not an object.**
   The example body is `[ { "id": 1, ... } ]`. So `GetRecordTemplate` must
   `httpclient.Get[[]RecordTemplate](...)` and return element `[0]`, erroring with
   `fmt.Errorf("record template %s not found", templateID)` when the slice is empty.
   Follow the pattern `api.CreatePatient` already uses for its single-element array.
2. `maxLength` and `defaultValue` are **strings**, not numbers.
3. Both endpoints need `system-admin`, which `auth login` does not grant by default.

`templateType` and keyword `type` are undocumented integer enums — print the raw ints,
invent no labels.

Commands:

- `record-templates list` — one line per template:
  `%-6d  %-40s  type=%d  %d keywords`, i.e. id, title, templateType, `len(Keywords)`.
- `record-templates get <templateId>` — header `%d  %s` (id, title), then
  `  Template type: %d`, `  Keywords: %d`, then one indented line per keyword
  (`  └─ %-5d %-30s type=%d`), **sorted by `Visual.Order`**.
- `record-templates keywords <templateId>` — the daily-use lookup. Flat table,
  **sorted by `Visual.Order`**, columns: `id`, `title`, `type`, `maxLength`,
  `defaultValue`, and a `*` marker in a trailing column when `WarnIfEmpty` is true
  (print a legend line `* = warn if empty` at the end). This is what you paste from to
  hand-build a `keywords[{id,value}]` payload for E1-3 and E1-4.

Sort with `sort.SliceStable` on a copy so `--json` still returns the API's own order.

## E1-2 · bookings create (P0, innovdr-proven)

`POST /v1/bookings` · scope `bookings:write`
Files: `internal/api/bookings.go`, `internal/commands/bookings.go`

Request (bold = what innovdr sends in production and therefore the required set):

```go
type CreateBookingRequest struct {
	UserID                    string   `json:"userId"`         // required
	ClinicID                  string   `json:"clinicId"`       // required
	PatientID                 string   `json:"patientId"`      // required
	BookingTypeID             int      `json:"bookingTypeId"`  // required
	PatientTypeID             int      `json:"patientTypeId"`  // required
	Date                      string   `json:"date"`           // required, YYYY-MM-DD
	StartTime                 string   `json:"startTime"`      // required, HH:MM
	EndTime                   string   `json:"endTime"`        // required, HH:MM
	InjuryNumber              string   `json:"injuryNumber,omitempty"`
	Note                      string   `json:"note,omitempty"`
	PreviousVisitID           string   `json:"previousVisitId,omitempty"`
	Fee                       *float64 `json:"fee,omitempty"`
	SMSReminder               *bool    `json:"smsReminder,omitempty"`
	SMSConfirmation           *bool    `json:"smsConfirmation,omitempty"`
	ActionCodes               []int    `json:"actionCodes,omitempty"`
	AllowOverlap              *bool    `json:"allowOverlap,omitempty"`
	ReservedForOrganizationID int      `json:"reservedForOrganizationId,omitempty"`
	ReservedForDepartmentID   int      `json:"reservedForDepartmentId,omitempty"`
}
```

`Fee`, `SMSReminder`, `SMSConfirmation`, `AllowOverlap` are pointers because `0` and
`false` are meaningful; set them only when `cmd.Flags().Changed(...)`.

Response:

```go
type Booking struct {
	ID                        string             `json:"id"`
	Date                      string             `json:"date"`
	StartTime                 string             `json:"startTime"`
	EndTime                   string             `json:"endTime"`
	Note                      string             `json:"note"`
	InjuryNumber              string             `json:"injuryNumber"`
	SMSReminder               bool               `json:"smsReminder"`
	ReservedForOrganizationID int                `json:"reservedForOrganizationId"`
	ReservedForDepartmentID   int                `json:"reservedForDepartmentId"`
	Phase                     int                `json:"phase"`
	Caregiver                 api.Caregiver      `json:"caregiver"`
	ActionCodes               []BookingActionRef `json:"actionCodes"`
	BookingType               api.Ref            `json:"bookingType"`
	PatientType               api.Ref            `json:"patientType"`
	Clinic                    BookingClinic      `json:"clinic"`
	Patient                   api.EmbeddedPatient `json:"patient"`
}

type BookingActionRef struct {
	ID       api.Loose `json:"id"`       // number on GET, string on POST — must be Loose
	CodeName string    `json:"codeName"`
}

type BookingClinic struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	HsaID       string      `json:"hsaId"`
	InvoiceInfo InvoiceInfo `json:"invoiceInfo"` // reuse the existing type in api/users.go
}
```

`bookingType.id` and `patientType.id` are numbers on `GET /v1/bookings` and strings on
the POST response — that is exactly why `api.Ref.ID` is `api.Loose`.

Flags: `--user-id`, `--patient-id`, `--booking-type-id`, `--patient-type-id`, `--date`,
`--start-time`, `--end-time` (all required); `--clinic-id` comes from `o.Clinic()`, not
a per-command flag. Optional: `--injury-number`, `--note`, `--previous-visit-id`,
`--fee`, `--sms-reminder`, `--sms-confirmation`, `--action-code` (`StringSlice`, parse
to `[]int`), `--allow-overlap`, `--reserved-for-organization-id`,
`--reserved-for-department-id`. Validate `--date` with `ValidateDate` and the two times
with `ValidateTime`.

Output: `Booking created: <id>`, then `  <date> <startTime>-<endTime>`,
`  Patient: <first> <last>`, `  Caregiver: <first> <last>`,
`  Booking type: <name>`.

## E1-3 · visits create + sign (P0, innovdr-proven)

`POST /v2/visits` (scope `self-service`) · `POST /v1/visits/:visitId/records/signatures`
(scope `record-signatures:write`)
Files: `internal/api/visits.go`, `internal/commands/visits.go`

```go
type CreateVisitRequest struct {
	BookingID string `json:"bookingId"` // required
	VisitDate string `json:"visitDate"` // required, "YYYY-MM-DD HH:MM:SS"
}

type Visit struct {
	ID              string              `json:"id"`
	ClinicID        string              `json:"clinicId"`
	VisitDate       string              `json:"visitDate"`
	BookingID       string              `json:"bookingId"`
	RecordTemplateID api.Loose          `json:"recordTemplateId"` // string on GET, absent on POST
	InjuryNumber    string              `json:"injuryNumber"`
	Payments        []api.Payment       `json:"payments"`
	Patient         api.EmbeddedPatient `json:"patient"`
	Caregiver       api.Caregiver       `json:"caregiver"`
}

type SignVisitRequest struct {
	UserID string `json:"userId"` // required
}

func CreateVisit(c *httpclient.Client, req CreateVisitRequest) (Visit, error)
func SignVisit(c *httpclient.Client, visitID string, req SignVisitRequest) error
```

`SignVisit` returns no body — use `httpclient.Post[struct{}]` and discard the value.

Commands:

- `visits create --booking-id <id> --visit-date "YYYY-MM-DD HH:MM:SS"` (both required,
  validate with `ValidateDateTime`). Output: `Visit created: <id>` then
  `  Visit date: <visitDate>`, `  Booking: <bookingId>`,
  `  Patient: <first> <last> (<personalNumber>)`.
- `visits sign <visitId> --user-id <id>` (`--user-id` required). Output:
  `Visit <visitId> signed`.

## E1-4 · records update (P0, innovdr-proven)

`PATCH /v1/visits/:visitId/records` · scope `patient:write`
Files: `internal/api/records.go`, `internal/commands/records.go`

```go
type UpdateRecordRequest struct {
	UserID         string        `json:"userId"`         // required
	Keywords       []api.Keyword `json:"keywords"`       // required (innovdr asserts non-nil)
	DiagnosisCodes []string      `json:"diagnosisCodes"` // required (send [] not null)
	ActionCodes    []string      `json:"actionCodes"`    // required (send [] not null)
}

// Note the ASYMMETRY: the request sends diagnosisCodes/actionCodes as []string,
// the response returns diagnoses/actions as objects with a `code` field.
type UpdatedRecord struct {
	UserID    string        `json:"userId"`
	Keywords  []api.Keyword `json:"keywords"`
	Diagnoses []CodeRef     `json:"diagnoses"`
	Actions   []CodeRef     `json:"actions"`
}

type CodeRef struct {
	Code string `json:"code"`
}
```

Initialise the slices to `[]api.Keyword{}` / `[]string{}` rather than leaving them nil,
so they serialise as `[]` not `null` — innovdr's production client asserts both are
non-nil.

Command `records update <visitId>`:

- `--user-id` (required)
- `--keyword id=value` — repeatable `StringArray`. Split on the **first** `=` only, so
  values may contain `=`. Error clearly on a non-integer id or a missing `=`.
  Get valid ids from `record-templates keywords <templateId>` (E1-1).
- `--diagnosis-code` — repeatable `StringArray` (e.g. `W02.03`)
- `--action-code` — repeatable `StringArray` (e.g. `PB003`)

Output: `Record updated for visit <visitId>`, then `  Keywords: <n>`,
`  Diagnoses: <comma-joined codes>`, `  Actions: <comma-joined codes>`.

## E1-5 · documents upload + document-types list (P0 / P1)

`POST /v1/clinics/:clinicId/documents` (multipart, scope `documents:write`) ·
`GET /v1/documentTypes` (scope `document-types:read`)
Files: `internal/api/documents.go`, `internal/commands/documents.go`,
`internal/commands/documenttypes.go`

Multipart form fields (from the collection and confirmed by innovdr's production
client, which sends the first four):

| field | type | required | notes |
|---|---|---|---|
| `file` | file | yes | the file part |
| `documentTypeId` | text | yes | from `document-types list` |
| `personalNumber` | text | yes | patient's personal number |
| `userId` | text | yes | uploading user |
| `createdAt` | text | no | `"YYYY-MM-DD HH:MM:SS"` |

```go
type UploadDocumentRequest struct {
	FilePath       string
	FileName       string // defaults to filepath.Base(FilePath)
	DocumentTypeID string
	PersonalNumber string
	UserID         string
	CreatedAt      string
}

func UploadDocument(c *httpclient.Client, clinicID string, req UploadDocumentRequest) error

type DocumentType struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	InReferral   bool   `json:"inreferral"`   // note: lowercase `inreferral` in the JSON
	Active       bool   `json:"active"`
	SignRequired bool   `json:"signRequired"`
}

func ListDocumentTypes(c *httpclient.Client) ([]DocumentType, error)
```

Build the body with `mime/multipart` into a `bytes.Buffer` and call
`httpclient.PostMultipart[struct{}](c, path, w.FormDataContentType(), &buf)` — that
helper already exists. Write the text fields with `w.WriteField`, the file with
`w.CreateFormFile("file", req.FileName)` + `io.Copy`, and **`w.Close()` before
sending** (a missing `Close` omits the trailing boundary and the API rejects it).
Upload returns no body. innovdr names files
`<personalNumber>_<yyyy-MM-dd_HHmm>_<name>.jpg`; keep `--file-name` as an override and
default to the basename.

Commands:

- `documents upload --file <path> --document-type-id <id> --personal-number <pn> --user-id <id> [--file-name <name>] [--created-at "..."]`.
  Clinic from `o.Clinic()`. Validate `--created-at` with `ValidateDateTime`. Output:
  `Uploaded <fileName> to clinic <clinicId>`.
- `document-types list` — `%-4d  %-40s  %s` with a trailing bracketed flag list, e.g.
  `[inactive]`, `[sign required]`, `[in-referral]`. Empty message
  `No document types found.`

## E1-6 · bookings list / list-by-clinic / update / cancel (P1)

Same files as E1-2.

**`bookings list` — `GET /v1/bookings`** (scope `bookings:read`). Query params:
`isBooked` (1/0), `userId`, `clinicId`, `bookingTypeId`, `fromDate`, `toDate`,
`reservedForOrganizationId`, `reservedForDepartmentId`, `offset`, `personalNumber`,
`arrivalStatus` (int). Here `clinicId` is an optional *filter*, so read `o.ClinicID`
directly — do **not** call `o.Clinic()`, which errors when unset. Response is
`[]Booking` (the E1-2 struct) with `user` instead of `caregiver`:

```go
User api.EmbeddedUser // add: {id, personalNumber, firstName, lastName, title}
```

Declare that small struct in your own `internal/api/bookings.go` as `BookingUser` —
don't add to `common.go`.

**`bookings list-by-clinic` — `GET /v1/clinics/:clinicId/bookings`** (scope
`self-service`, clinic from `o.Clinic()`). Query: `personalNumber`, `date`, `fromDate`,
`toDate`, `arrived` (1/0), `selfService` (bool), `bookingTypeId`, `booked` (bool),
`userId`, `limit`, `offset`, `reservedForOrganizationId`, `reservedForDepartmentId`.
Use `AddPagination`/`SetPagination`. This endpoint returns a **different shape** —
`bookingType` and `bookedPatientType` are plain **strings**, not objects:

```go
type ClinicBooking struct {
	ID                string              `json:"id"`
	ClinicID          string              `json:"clinicId"`
	Date              string              `json:"date"`
	StartTime         string              `json:"startTime"`
	EndTime           string              `json:"endTime"`
	Arrived           int                 `json:"arrived"`
	ActionCodes       []BookingActionRef  `json:"actionCodes"`
	BookingDescription string             `json:"bookingDescription"`
	BookingType       string              `json:"bookingType"`       // string here!
	BookedPatientType string              `json:"bookedPatientType"` // string here!
	InjuryNumber      string              `json:"injuryNumber"`
	Payments          []api.Payment       `json:"payments"`
	Patient           api.EmbeddedPatient `json:"patient"`
	Caregiver         api.Caregiver       `json:"caregiver"`
}
```

`api.EmbeddedPatient` deliberately **omits `freeCard`**, because this endpoint returns
it as an object and `PATCH /v1/bookings/:id` returns it as `[]` — one struct cannot
decode both. If you need it, add a `FreeCard json.RawMessage` field on
`ClinicBooking`, not on the shared type.

**`bookings update <bookingId>` — `PATCH /v1/bookings/:bookingId`** (scope
`bookings:write`). Body — every field optional, all `omitempty`:

```go
type UpdateBookingRequest struct {
	RequestedByUserID         string      `json:"requestedByUserId,omitempty"`
	BookingTypeID             int         `json:"bookingTypeId,omitempty"`
	Caregiver                 *api.IDRef  `json:"caregiver,omitempty"`
	ReservedForOrganizationID int         `json:"reservedForOrganizationId,omitempty"`
	ReservedForDepartmentID   int         `json:"reservedForDepartmentId,omitempty"`
	PatientID                 string      `json:"patientId,omitempty"`
	Date                      string      `json:"date,omitempty"`
	StartTime                 string      `json:"startTime,omitempty"`
	EndTime                   string      `json:"endTime,omitempty"`
	Note                      string      `json:"note,omitempty"`
	InjuryNumber              string      `json:"injuryNumber,omitempty"`
	ActionCodes               []int       `json:"actionCodes,omitempty"`
}
```

Response decodes as `ClinicBooking`. Flags mirror the field names
(`--requested-by-user-id`, `--caregiver-id`, ...). Error if no flag was set at all.

**`bookings cancel <bookingId>` — `DELETE /v1/bookings/:bookingId`** (scope
`bookings:write`). This DELETE **requires a body**: `{"userId": "..."}`. Use
`httpclient.Delete[struct{}](c, path, DeleteBookingRequest{UserID: userID})` with
`--user-id` required. Output: `Booking <bookingId> cancelled`.

## E1-7 · visits list (P1)

`GET /v1/clinics/:clinicId/visits` · scope `self-service` · same files as E1-3.
Clinic from `o.Clinic()`. Query: `personalNumber`, `date`, `fromDate`, `toDate`,
`userId`, `recordTemplateId`, `isSigned` (1/0 → `SetBool01`). Response `[]Visit`.

Output: `%s  %s  %s %s` = id, visitDate, patient first+last, then an indented
`  └─ caregiver <title> <first> <last>` line. Empty: `No visits found.`

## E1-8 · notes create (P1)

`POST /v1/notes` · scope `patient:write`
Files: `internal/api/notes.go`, `internal/commands/notes.go`

```go
type CreateNoteRequest struct {
	PatientID        string `json:"patientId"`        // required
	ClinicID         string `json:"clinicId"`         // required, from o.Clinic()
	PatientTypeID    int    `json:"patientTypeId"`    // required
	UserID           string `json:"userId"`           // required
	RecordTemplateID int    `json:"recordTemplateId"` // required
	CreatedAt        string `json:"createdAt"`        // "YYYY-MM-DD HH:MM:SS"
}

type Note struct {
	ID               string `json:"id"`
	PatientID        string `json:"patientId"`
	ClinicID         string `json:"clinicId"`
	PatientTypeID    int    `json:"patientTypeId"`
	UserID           string `json:"userId"`
	RecordTemplateID int    `json:"recordTemplateId"`
	CreatedAt        string `json:"createdAt"`
}
```

`notes create --patient-id --patient-type-id --user-id --record-template-id [--created-at]`.
Validate `--created-at` with `ValidateDateTime`. Output: `Note created: <id>` then
`  Patient: <patientId>`, `  Template: <recordTemplateId>`, `  Created: <createdAt>`.

## E1-9 · records create-external (P2)

`POST /v1/externalRecords` · scope `system-admin` · same files as E1-4.

```go
type CreateExternalRecordRequest struct {
	PatientIdentity   string        `json:"patientIdentity"`   // required, patient UUID
	AuthorIdentity    int           `json:"authorIdentity"`    // required
	RecordTemplateID  string        `json:"recordTemplateId"`  // required — STRING in the request
	Source            string        `json:"source"`
	QuestionnaireName string        `json:"questionnaireName"`
	Keywords          []api.Keyword `json:"keywords"`
	ActionCodes       []string      `json:"actionCodes"`
	DiagnosisCodes    []string      `json:"diagnosisCodes"`
}

type ExternalRecord struct {
	UUID              string                 `json:"uuid"`
	PatientIdentity   string                 `json:"patientIdentity"`
	AuthorIdentity    int                    `json:"authorIdentity"`
	RecordTemplateID  api.Loose              `json:"recordTemplateId"` // NUMBER in the response
	QuestionnaireName string                 `json:"questionnaireName"`
	Source            string                 `json:"source"`
	Keywords          []ExternalRecordKeyword `json:"keywords"`
	ActionCodes       []string               `json:"actionCodes"`
	DiagnosisCodes    []string               `json:"diagnosisCodes"`
}

type ExternalRecordKeyword struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Value string `json:"value"`
	Type  int    `json:"type"`
}
```

`recordTemplateId` is a string going out and a number coming back — hence `api.Loose`
on the response only. Reuse the `--keyword id=value` parser from E1-4.
Output: `External record created: <uuid>` plus a short summary.

---

# Engineer 2 — the read / reference surface

Order: E2-1 → E2-2 → E2-3 first (these are what people reach for), then E2-4 → E2-6.

Only you edit `internal/api/users.go`, `internal/api/patients.go`,
`internal/api/patients_types.go`, `internal/commands/users.go` and
`internal/commands/patients.go`.

## E2-1 · clinics list + health (P1 / P2)

`GET /v1/clinics` (scope `clinics:read`) · `GET /v1/health-check` (no scope)
Files: `internal/api/clinics.go`, `internal/commands/clinics.go`,
`internal/api/health.go`, `internal/commands/health.go`

```go
type Clinic struct {
	ID                 string      `json:"id"`
	Name               string      `json:"name"`
	HsaID              string      `json:"hsaId"`
	CompanyHsaID       string      `json:"companyHsaId"`
	OrganisationNumber string      `json:"organisationNumber"`
	Address            string      `json:"address"`
	PostNumber         string      `json:"postNumber"`
	PostAddress        string      `json:"postAddress"`
	InvoiceInfo        InvoiceInfo `json:"invoiceInfo"`
}

type Health struct {
	Status string `json:"status"`
}
```

`InvoiceInfo` already exists in `internal/api/users.go` with only `companyName`.
**Extend that existing type** (you own the file) rather than declaring a second one —
engineer 1 references it too:

```go
type InvoiceInfo struct {
	CompanyName        string `json:"companyName"`
	OrganisationNumber string `json:"organisationNumber,omitempty"`
	CompanyTelephone   string `json:"companyTelephone,omitempty"`
	CompanyAddress     string `json:"companyAddress,omitempty"`
	CompanyPost        string `json:"companyPost,omitempty"`
	CompanyPostArea    string `json:"companyPostArea,omitempty"`
	BankGiro           string `json:"bankGiro,omitempty"`
	PostGiro           string `json:"postGiro,omitempty"`
}
```

- `clinics list [--hsa-id] [--company-hsa-id] [--organisation-number]` —
  `%s  %-40s  %s` = id, name, hsaId; indented `  └─ <address>, <postNumber> <postAddress>`.
  Empty: `No clinics found.`
- `health` — a top-level command with no subcommands (`Use: "health"`,
  `Short: "Check API health"`). Output: `status: <status>`. Return a non-nil error when
  the status is not `ok`/`OK` so it is usable in a shell check.

## E2-2 · users list / get / permissions / list-by-clinic (P1)

`GET /v1/users` · `GET /v1/users/:userId` · `GET /v1/users/:uuid/permissions` ·
`GET /v1/clinics/:clinicId/users` — all scope `users:read`
Files: `internal/api/users.go`, `internal/commands/users.go` (both extended)

Leave `users search` and `api.SearchUsers` exactly as they are.

**Trap — the list and by-id responses differ in two field names:**

| list (`GET /v1/users`) | by id (`GET /v1/users/:userId`) |
|---|---|
| `phoneNumber` | `telephoneNumber` |
| `settings.defaultCostCentreId` (British) | `settings.defaultCostCenterId` (American) |

Do not share one struct across both blindly. Either give the existing `User` struct
both spellings (all `omitempty`, and pick whichever is non-empty when printing) or
declare a separate `UserDetail`. Prefer adding both tags to `User` — fewer types.

```go
type UserSettings struct {
	DefaultCostCentreID     string  `json:"defaultCostCentreId,omitempty"`
	DefaultCostCenterID     string  `json:"defaultCostCenterId,omitempty"`
	DefaultRecordTemplateID *string `json:"defaultRecordTemplateId"` // nullable
	DefaultPatientTypeID    string  `json:"defaultPatientTypeId"`
}

// GET /v1/clinics/:clinicId/users returns only these three fields.
type ClinicUser struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}
```

`GET /v1/users/:uuid/permissions` returns a **map keyed by clinic UUID**, whose values
are maps of permission name → string value (`"1"`/`"0"`), with ~44 keys per clinic:

```go
type Permissions map[string]map[string]string

func GetUserPermissions(c *httpclient.Client, uuid string) (Permissions, error)
```

Commands:

- `users list [--first-name] [--last-name] [--personal-number] [--hsa-id] [--id]` —
  same per-user output as `users search`. Empty: `No users found.`
- `users get <userId>` — header `%s  %s %s  (%s)` = id, first, last, personalNumber,
  then indented `HSA ID`, `Email`, `Phone`, `Last logged in`, then one
  `  └─ <clinicId>  <clinicName>` per clinic, then a `Settings:` block.
- `users permissions <uuid>` — for each clinic UUID print `<clinicId>` as a header,
  then only the permissions whose value is truthy (`"1"`), two-space indented, sorted
  alphabetically (map iteration order is random — sort the keys or the output is
  non-deterministic). Add `--all` to include the falsy ones.
- `users list-by-clinic [--is-caregiver]` — clinic from `o.Clinic()`;
  `--is-caregiver` is a bool query param, emit only when `Changed` (`SetBool`).
  Output `%s  %s %s`. Empty: `No users found for clinic <clinicId>.`

## E2-3 · patients: list / update / create-without-ssn / national-registry / set-current

Files: `internal/api/patients.go`, `internal/api/patients_types.go`,
`internal/commands/patients.go` (all extended)

Leave `patients search` and `patients create` as they are.

**`patients list` — `GET /v1/patients`** (scope `self-service`, P2). Deprecated
upstream ("please use GET /v2/patients instead") — put that in the command's `Short`
and add `Deprecated: "use \"patients search\" instead"` on the cobra command so cobra
prints the warning. It "fetches information for a specific patient" and returns a
**single object, not an array**. No query params are documented; send
`personalNumber` as a query param, mirroring the existing v2 call, and verify against
integration. Response shape:

```go
type PatientV1 struct {
	ID                  string         `json:"id"`
	PersonalNumber      string         `json:"personalNumber"`
	BirthDate           string         `json:"birthDate"`
	Gender              string         `json:"gender"`
	FirstName           string         `json:"firstName"`
	LastName            string         `json:"lastName"`
	Address             PatientAddress `json:"address"`       // reuse existing type
	Nationality         string         `json:"nationality"`
	Email               string         `json:"email"`
	MobilePhoneNumber   string         `json:"mobilePhoneNumber"`
	WorkPhoneNumber     string         `json:"workPhoneNumber"`
	HomePhoneNumber     string         `json:"homePhoneNumber"`
	ListedClinic        ListedClinic   `json:"listedClinic"`  // reuse existing type
	InterpreterNeeded   bool           `json:"interpreterNeeded"`
	InterpreterLanguage string         `json:"interpreterLanguage"`
	PatientType         PatientType    `json:"patientType"`   // reuse existing type
}
```

**`patients update <personalNumber>` — `PATCH /v1/patients/:personalNumber`** (scope
`patient:write`, P1). All fields optional, no response body:

```go
type UpdatePatientRequest struct {
	ListInfo     *CreatePatientListInfo `json:"listInfo,omitempty"` // reuse existing
	PatientType  string `json:"patientType,omitempty"`
	Email        string `json:"email,omitempty"`
	MobileNumber string `json:"mobileNumber,omitempty"`
	FirstName    string `json:"firstName,omitempty"`
	LastName     string `json:"lastName,omitempty"`
	SpokenName   string `json:"spokenName,omitempty"`
	Address      string `json:"address,omitempty"`
	CoAddress    string `json:"coAddress,omitempty"`
	City         string `json:"city,omitempty"`
	ZipCode      string `json:"zipCode,omitempty"`
	StateCode    string `json:"stateCode,omitempty"`
	CountyCode   string `json:"countyCode,omitempty"`
}
```

Send `listInfo` only when at least one of `--clinic-name`/`--clinic-hsa-id`/
`--county-name` was given (that is why it is a pointer). Error if no flag was set.
`httpclient.Patch[struct{}]`. Output: `Patient <personalNumber> updated`.

**`patients create-without-ssn` — `POST /v1/patient/withoutSocialSecurityNumber`**
(scope `patient:write`, P1). Response is an **array** — take `[0]`, like
`api.CreatePatient` does.

```go
type CreatePatientWithoutSSNRequest struct {
	IdentificationNumber string                `json:"identificationNumber"` // required
	FirstName            string                `json:"firstName"`            // required
	LastName             string                `json:"lastName"`             // required
	BirthDate            string                `json:"birthDate"`            // required, YYYY-MM-DD
	Gender               string                `json:"gender"`               // required, "M"/"K"
	PatientTypeID        int                   `json:"patientTypeId"`        // required
	Contact              CreatedPatientContact `json:"contact"`              // reuse existing
	ListInfo             CreatePatientListInfo `json:"listInfo"`             // reuse existing
}

type PatientWithoutSSN struct {
	ID                   string                `json:"id"`
	IdentificationNumber string                `json:"identificationNumber"`
	FirstName            string                `json:"firstName"`
	LastName             string                `json:"lastName"`
	BirthDate            string                `json:"birthDate"`
	Gender               string                `json:"gender"`
	Contact              CreatedPatientContact `json:"contact"`
	ListInfo             CreatedPatientListInfo `json:"listInfo"`
	PatientType          PatientTypeRef        `json:"patientType"` // {id int, name string}
}
```

Flags: `--identification-number`, `--first-name`, `--last-name`, `--birth-date`
(`ValidateDate`), `--gender`, `--patient-type-id`, `--clinic-name`, `--clinic-hsa-id`,
`--county-name` (required); `--street-name`, `--zip-code`, `--city`, `--co-address`,
`--mobile`, `--email`, `--work-phone`, `--home-phone` (optional contact block).

**`patients national-registry <personalNumber>` — `GET /v1/nationalRegistry/:personalNumber`**
(scope `patient:read`, P1). Single object; `unRegisterCauseCode` is nullable:

```go
type NationalRegistryPerson struct {
	UnRegisterCauseCode *string `json:"unRegisterCauseCode"`
	FirstName           string  `json:"firstName"`
	LastName            string  `json:"lastName"`
	SpokenName          string  `json:"spokenName"`
	MaritialStatus      string  `json:"maritialStatus"` // API's own spelling — keep it
	CoAddress           string  `json:"coAddress"`
	Address             string  `json:"address"`
	ZipCode             string  `json:"zipCode"`
	City                string  `json:"city"`
	StateCode           string  `json:"stateCode"`
	CountyCode          string  `json:"countyCode"`
	Parish              string  `json:"parish"`
	Deceased            string  `json:"deceased"`
}
```

Output: `%s %s` header then indented `Address`, `Parish`, `County`, `State`, and only
print `Deceased` / `Unregistered` when non-empty.

**`patients set-current <personalNumber>` — `PUT /v1/currentUser/currentPatient`**
(scope `patient:write`, P2). Body `{"personalNumber": "...", "showWarning": true}`;
`--show-warning` bool defaulting to `true`. No response body — use
`httpclient.Put[struct{}]`. Output: `Current patient set to <personalNumber>`.

## E2-4 · action-codes + payment-methods (P1)

`GET /v1/actionCodes` (scope `actioncodes:read`) ·
`GET /v1/clinics/:clinicId/paymentMethods` (scope `self-service`)
Files: `internal/api/actioncodes.go`, `internal/commands/actioncodes.go`,
`internal/api/paymentmethods.go`, `internal/commands/paymentmethods.go`

**Name collision:** `internal/api/bookingtypes.go` already declares
`type ActionCode struct { ID int; Name string }`, which is *not* this endpoint's shape.
Name the new one **`ActionCodeDetail`**.

```go
type ActionCodeDetail struct {
	ID              int       `json:"id"`
	ArticleNo       api.Loose `json:"articleNo"`   // null in the example — Loose absorbs it
	CodeName        string    `json:"codeName"`
	CodeDescription string    `json:"codeDescription"`
	Account         int       `json:"account"`
	Fee             api.Loose `json:"fee"`          // string in the example
	Compensation    api.Loose `json:"compensation"` // string in the example
	Active          int       `json:"active"`       // 1/0, not a bool
}

type PaymentMethod struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	DirectPayment bool   `json:"directPayment"`
	Active        bool   `json:"active"`
	AssetsAccount string `json:"assetsAccount"`
}
```

Inside package `api` you refer to these as `Loose`, not `api.Loose`.

- `action-codes list` — `%-6d  %-10s  %-40s  fee=%s` = id, codeName, codeDescription,
  fee; append ` [inactive]` when `Active != 1`. Empty: `No action codes found.`
- `payment-methods list` — clinic from `o.Clinic()`. `%-4s  %-30s  %s` = id, name,
  assetsAccount; append ` [direct]` / ` [inactive]`.
  Empty: `No payment methods found.`

## E2-5 · invoices (P1 / P2)

`GET /v1/clinics/:clinicId/invoices` · `GET /v1/clinics/:clinicId/creditInvoices` ·
`POST /v1/clinics/:clinicId/invoices` · `PATCH /v1/clinics/:clinicId/invoices/:invoiceId`
— all scope `self-service`, clinic from `o.Clinic()`
Files: `internal/api/invoices.go`, `internal/commands/invoices.go`

Every one of the four returns the same `Invoice` shape. **All the money fields are
strings in the invoice examples but numbers on bookings/visits — use `Loose`.**

```go
type Invoice struct {
	InvoiceID     string           `json:"invoiceId"`
	InvoiceNumber string           `json:"invoiceNumber"`
	InvoiceStatus string           `json:"invoiceStatus"`
	Currency      string           `json:"currency"`
	NetAmount     Loose            `json:"netAmount"`
	TotalAmount   Loose            `json:"totalAmount"`
	VAT           Loose            `json:"VAT"`
	VATPercentage Loose            `json:"VATPercentage"`
	ClinicName    string           `json:"clinicName"`
	PaidDate      *string          `json:"paidDate"`   // null in the examples
	ExpiryDate    string           `json:"expiryDate"`
	CreatedAt     string           `json:"createdAt"`
	Payments      []Payment        `json:"payments"`
	Caregiver     Caregiver        `json:"caregiver"`
	Patient       EmbeddedPatient  `json:"patient"`
	Recipient     InvoiceRecipient `json:"recipient"`
}

type InvoiceRecipient struct {
	Type           string `json:"type"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	BillingAddress string `json:"billingAddress"`
	BillingZipCode string `json:"billingZipCode"`
	BillingCity    string `json:"billingCity"`
}

type CreateInvoiceRequest struct {
	VisitID      string `json:"visitId"`      // required
	TotalPayment string `json:"totalPayment"` // required — a STRING, e.g. "123"
}

type UpdateInvoiceRequest struct {
	TotalPayment  string `json:"totalPayment,omitempty"`
	PaymentMethod string `json:"paymentMethod,omitempty"` // a STRING, e.g. "1"
}
```

`totalPayment` and `paymentMethod` are strings on the wire. Take them as string flags
and pass them through verbatim — do not parse to a number and re-serialise.

**Verify the list cardinality against integration.** The collection shows the two GET
examples as a single object, but `POST` returns an array of the same shape, so the GETs
are almost certainly arrays too and a docs shortcut is the likelier explanation.
Implement as `[]Invoice`; if integration returns an object, switch those two to a
single `Invoice` and note it in the command's `Short`.

- `invoices list` / `invoices list-credit` — query `personalNumber`, `date`,
  `fromDate`, `toDate`, `isExported` (1/0 → `SetBool01`); `list` also has `status`
  (int). Output `%-12s  %-10s  %10s %s  %s` = invoiceNumber, invoiceStatus,
  totalAmount, currency, patient name; indented `  └─ created <createdAt>  expires <expiryDate>`.
  Empty: `No invoices found.`
- `invoices create --visit-id <id> --total-payment <amount>` — response is an array,
  take `[0]`. Output: `Invoice created: <invoiceNumber> (<invoiceId>)`.
- `invoices update <invoiceId> [--total-payment] [--payment-method]` — response is a
  single object. Error if neither flag was set.
  Output: `Invoice <invoiceNumber> updated` plus status and total.

## E2-6 · organizations (P2)

`GET /v1/organizations` (scope `organization:read`) ·
`POST /v1/organization/organizations` · `PUT /v1/organizations/:organizationId/patient`
(both scope `organization:write`)
Files: `internal/api/organizations.go`, `internal/commands/organizations.go`

Note the POST path really is singular `/v1/organization/organizations` while the others
are plural — not a typo.

```go
type Organization struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Active      int          `json:"active"` // 1/0, not a bool
	Departments []Department `json:"departments"`
}

type Department struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Parent string `json:"parent"`
}

// POST body is a batch: {"rows": [ ... ]}
type CreateOrganizationsRequest struct {
	Rows []OrganizationRow `json:"rows"`
}

type OrganizationRow struct {
	Level1               string `json:"level_1"` // required — the top-level company name
	Level2               string `json:"level_2"` // levels 2..7 are nullable sub-levels
	Level3               string `json:"level_3"`
	Level4               string `json:"level_4"`
	Level5               string `json:"level_5"`
	Level6               string `json:"level_6"`
	Level7               string `json:"level_7"`
	OrganizationNumber   string `json:"organizationNumber"`
	CompanyText          string `json:"companyText"`
	Email                string `json:"email"`
	StreetAddress        string `json:"streetAddress"`
	ZipCode              string `json:"zipCode"`
	City                 string `json:"city"`
	InvoiceEmail         string `json:"invoiceEmail"`
	InvoiceStreetAddress string `json:"invoiceStreetAddress"`
	InvoiceZipCode       string `json:"invoiceZipCode"`
	InvoiceCity          string `json:"invoiceCity"`
}

type CreateOrganizationsResponse struct {
	UUID string `json:"uuid"`
}

type AddPatientToOrganizationRequest struct {
	PatientID    string `json:"patientId"`    // required
	DepartmentID string `json:"departmentId"` // required
}
```

The `level_*` fields are the org hierarchy: each row names a path from company down to
sub-department, with unused levels `null`. The example sends `null`, so **do not** use
`omitempty` on them — send empty strings, or make them `*string` and set only the ones
given. Prefer `*string` to match the documented `null`.

Because the request is a batch of many-field rows, take it from a file rather than 17
flags: `organizations create --file <path>` reading a JSON array of rows (or a full
`{"rows":[...]}` object — accept both), plus a `--level-1` shorthand for the
single-row case. Say so in the `Long` help.

- `organizations list` — `%s  %-40s` = id, name (` [inactive]` when `Active != 1`),
  then one `  └─ %s  %s` per department. Empty: `No organizations found.`
- `organizations create --file <path>` — output `Organizations created: <uuid>` and
  `  <n> rows`.
- `organizations add-patient <organizationId> --patient-id <id> --department-id <id>` —
  `httpclient.Put[struct{}]`, no response body.
  Output: `Patient <patientId> added to department <departmentId>`.

---

## Definition of done (both engineers)

```sh
go build ./... && go vet ./... && gofmt -l .     # gofmt must print nothing
go run ./cmd/webdoc --help
go run ./cmd/webdoc <your-group> --help
go run ./cmd/webdoc <your-group> <verb> --help
```

Plus: one real call per command against the integration environment where a token with
the right scope is available, checked both with and without `--json`. Where a response
shape is flagged above as needing verification (invoice list cardinality,
`patients list` query parameter), record what the API actually returned in your PR
description.
