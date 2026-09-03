package commands

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

// Layouts Webdoc uses on the wire.
const (
	DateLayout     = "2006-01-02"
	TimeLayout     = "15:04"
	DateTimeLayout = "2006-01-02 15:04:05"
)

// ValidateDate returns value unchanged, erroring if it is set but not YYYY-MM-DD.
// An empty value is allowed so it can be used on optional filter flags.
func ValidateDate(flag, value string) (string, error) {
	return validate(flag, value, DateLayout, "YYYY-MM-DD")
}

// ValidateTime checks HH:MM.
func ValidateTime(flag, value string) (string, error) {
	return validate(flag, value, TimeLayout, "HH:MM")
}

// ValidateDateTime checks "YYYY-MM-DD HH:MM:SS". Fractional seconds are also
// accepted - time.Parse allows them after the seconds field even though the
// layout omits them - so the collection's "11:37:25.123" examples pass.
func ValidateDateTime(flag, value string) (string, error) {
	return validate(flag, value, DateTimeLayout, "YYYY-MM-DD HH:MM:SS")
}

func validate(flag, value, layout, human string) (string, error) {
	if value == "" {
		return "", nil
	}
	if _, err := time.Parse(layout, value); err != nil {
		return "", fmt.Errorf("--%s must be %s, got %q", flag, human, value)
	}
	return value, nil
}

// SetStr adds key=value to q, skipping empty values.
func SetStr(q url.Values, key, value string) {
	if value != "" {
		q.Set(key, value)
	}
}

// SetInt adds key=value to q, skipping zero.
func SetInt(q url.Values, key string, value int) {
	if value != 0 {
		q.Set(key, strconv.Itoa(value))
	}
}

// SetBool adds key=true/false to q only when the user actually passed the flag.
func SetBool(cmd *cobra.Command, q url.Values, key, flag string) {
	if !cmd.Flags().Changed(flag) {
		return
	}
	v, _ := cmd.Flags().GetBool(flag)
	q.Set(key, strconv.FormatBool(v))
}

// SetBool01 is SetBool for the endpoints that want 1/0 rather than true/false
// (arrived, isBooked, isExported, isSigned, arrivalStatus).
func SetBool01(cmd *cobra.Command, q url.Values, key, flag string) {
	if !cmd.Flags().Changed(flag) {
		return
	}
	v, _ := cmd.Flags().GetBool(flag)
	if v {
		q.Set(key, "1")
	} else {
		q.Set(key, "0")
	}
}

// AddPagination registers the standard --limit/--offset pair.
func AddPagination(cmd *cobra.Command) {
	cmd.Flags().Int("limit", 0, "Max number of results")
	cmd.Flags().Int("offset", 0, "Number of results to skip")
}

// SetPagination copies --limit/--offset into q when set.
func SetPagination(cmd *cobra.Command, q url.Values) {
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	SetInt(q, "limit", limit)
	SetInt(q, "offset", offset)
}
