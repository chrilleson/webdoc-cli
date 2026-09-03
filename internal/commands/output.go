package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/chrilleson/webdoc-cli/internal/httpclient"
)

// Emit is the single output path for every command. With --json it prints v as
// indented JSON; otherwise it calls render, which does the printf formatting.
//
//	return commands.Emit(o, bookings, func() {
//		for _, b := range bookings {
//			fmt.Printf("%s  %s %s-%s\n", b.ID, b.Date, b.StartTime, b.EndTime)
//		}
//	})
func Emit(o *Options, v any, render func()) error {
	if o.JSON {
		return PrintJSON(v)
	}
	render()
	return nil
}

// EmitEmpty is Emit for list endpoints: with --json an empty slice still prints
// as `[]`, otherwise the human-readable "none found" message is used.
//
//	return commands.EmitEmpty(o, bookings, len(bookings) == 0, "No bookings found.", func() { ... })
func EmitEmpty(o *Options, v any, empty bool, message string, render func()) error {
	if o.JSON {
		return PrintJSON(v)
	}
	if empty {
		fmt.Println(message)
		return nil
	}
	render()
	return nil
}

func PrintJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// ScopeHint wraps a 401/403 with a reminder that the endpoint needs a scope the
// default `auth login --scope self-service` does not grant. Use it on endpoints
// whose collection docs name a scope, e.g. ScopeHint(err, "system-admin").
func ScopeHint(err error, scope string) error {
	var apiErr *httpclient.APIError
	if errors.As(err, &apiErr) &&
		(apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden) {
		return fmt.Errorf("%w\nthis endpoint requires the %q scope - run `webdoc auth login --scope %s ...`", err, scope, scope)
	}
	return err
}
