# Webdoc API coverage

All 41 endpoints from the published Postman collection (docs.webdoc.com).

- **Status** — `done` = already shipped, `new` = to be implemented, `n/a` = handled outside the command tree.
- **Priority** — `P0` = proven in production by `innovdr-backend`, or a hard dependency of that path. `P1` = core CRUD / reference data. `P2` = nice-to-have.
- **Scope** — the OAuth2 scope the collection documents for the endpoint. `auth login` defaults to `--scope self-service`, so anything else needs an explicit `--scope`.

| # | Method | Path | CLI command | Status | Prio | Owner | Scope |
|---|---|---|---|---|---|---|---|
| 1 | POST | `https://auth.atlan.se/oauth/token` | `auth login` | done | P0 | — | — |
| 2 | GET | `/v1/health-check` | `health` | new | P2 | eng2 | none |
| 3 | GET | `/v1/actionCodes` | `action-codes list` | new | P1 | eng2 | `actioncodes:read` |
| 4 | GET | `/v1/clinics` | `clinics list` | new | P1 | eng2 | `clinics:read` |
| 5 | GET | `/v1/bookingTypes` | `booking-types list` | done | P0 | — | `self-service` |
| 6 | POST | `/v1/bookings` | `bookings create` | new | **P0** | eng1 | `bookings:write` |
| 7 | GET | `/v1/bookings` | `bookings list` | new | P1 | eng1 | `bookings:read` |
| 8 | GET | `/v1/clinics/:clinicId/bookings` | `bookings list-by-clinic` | new | P1 | eng1 | `self-service` |
| 9 | PATCH | `/v1/bookings/:bookingId` | `bookings update <bookingId>` | new | P1 | eng1 | `bookings:write` |
| 10 | DELETE | `/v1/bookings/:bookingId` | `bookings cancel <bookingId>` | new | P1 | eng1 | `bookings:write` |
| 11 | POST | `/v2/visits` | `visits create` | new | **P0** | eng1 | `self-service` |
| 12 | GET | `/v1/clinics/:clinicId/visits` | `visits list` | new | P1 | eng1 | `self-service` |
| 13 | POST | `/v1/visits/:visitId/records/signatures` | `visits sign <visitId>` | new | **P0** | eng1 | `record-signatures:write` |
| 14 | PATCH | `/v1/visits/:visitId/records` | `records update <visitId>` | new | **P0** | eng1 | `patient:write` |
| 15 | POST | `/v1/externalRecords` | `records create-external` | new | P2 | eng1 | `system-admin` |
| 16 | GET | `/v1/recordTemplates` | `record-templates list` | new | **P0** | eng1 | `system-admin` |
| 17 | GET | `/v1/recordTemplates/:templateId` | `record-templates get <templateId>` + `record-templates keywords <templateId>` | new | **P0** | eng1 | `system-admin` |
| 18 | POST | `/v1/clinics/:clinicId/documents` | `documents upload` | new | **P0** | eng1 | `documents:write` |
| 19 | GET | `/v1/documentTypes` | `document-types list` | new | P1 | eng1 | `document-types:read` |
| 20 | POST | `/v1/notes` | `notes create` | new | P1 | eng1 | `patient:write` |
| 21 | GET | `/v2/patients` | `patients search` | done | P0 | — | `patient:read` |
| 22 | POST | `/v1/patients` | `patients create` | done | P0 | — | `patient:write` |
| 23 | GET | `/v1/patients` | `patients list` (deprecated upstream) | new | P2 | eng2 | `self-service` |
| 24 | PATCH | `/v1/patients/:personalNumber` | `patients update <personalNumber>` | new | P1 | eng2 | `patient:write` |
| 25 | POST | `/v1/patient/withoutSocialSecurityNumber` | `patients create-without-ssn` | new | P1 | eng2 | `patient:write` |
| 26 | GET | `/v1/nationalRegistry/:personalNumber` | `patients national-registry <personalNumber>` | new | P1 | eng2 | `patient:read` |
| 27 | PUT | `/v1/currentUser/currentPatient` | `patients set-current <personalNumber>` | new | P2 | eng2 | `patient:write` |
| 28 | GET | `/v1/patientTypes` | `patient-types list` | done | P1 | — | `patient-types:read` |
| 29 | GET | `/v1/patientTypes/:id` | `patient-types get <id>` | done | P1 | — | `patient-types:read` |
| 30 | GET | `/v1/users` | `users search` (by personalNumber) | done | P0 | — | `users:read` |
| 31 | GET | `/v1/users` | `users list` (other filters) | new | P1 | eng2 | `users:read` |
| 32 | GET | `/v1/users/:userId` | `users get <userId>` | new | P1 | eng2 | `users:read` |
| 33 | GET | `/v1/users/:uuid/permissions` | `users permissions <uuid>` | new | P1 | eng2 | `users:read` |
| 34 | GET | `/v1/clinics/:clinicId/users` | `users list-by-clinic` | new | P1 | eng2 | `users:read` |
| 35 | GET | `/v1/organizations` | `organizations list` | new | P2 | eng2 | `organization:read` |
| 36 | POST | `/v1/organization/organizations` | `organizations create` | new | P2 | eng2 | `organization:write` |
| 37 | PUT | `/v1/organizations/:organizationId/patient` | `organizations add-patient <organizationId>` | new | P2 | eng2 | `organization:write` |
| 38 | GET | `/v1/clinics/:clinicId/invoices` | `invoices list` | new | P1 | eng2 | `self-service` |
| 39 | GET | `/v1/clinics/:clinicId/creditInvoices` | `invoices list-credit` | new | P2 | eng2 | `self-service` |
| 40 | POST | `/v1/clinics/:clinicId/invoices` | `invoices create` | new | P1 | eng2 | `self-service` |
| 41 | PATCH | `/v1/clinics/:clinicId/invoices/:invoiceId` | `invoices update <invoiceId>` | new | P1 | eng2 | `self-service` |
| 42 | GET | `/v1/clinics/:clinicId/paymentMethods` | `payment-methods list` | new | P1 | eng2 | `self-service` |

Row 30/31 are the same endpoint reached by two commands (`search` already exists and is
kept for backwards compatibility); that is why the table has 42 rows for 41 endpoints.

## Scope groups you will need while testing

```
webdoc auth login --client-id … --client-secret … --scope "self-service"
webdoc auth login --client-id … --client-secret … --scope "system-admin"          # record templates, external records
webdoc auth login --client-id … --client-secret … --scope "bookings:write patient:write documents:write record-signatures:write"
webdoc auth login --client-id … --client-secret … --scope "users:read clinics:read patient-types:read actioncodes:read organization:read document-types:read payment-methods:read"
```

Scopes are per-token, so switching workstreams may mean re-running `auth login`.
