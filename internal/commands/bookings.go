package commands

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/chrilleson/webdoc-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewBookingsCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bookings",
		Short: "Manage bookings",
	}
	cmd.AddCommand(
		newBookingsListCmd(o),
		newBookingsListByClinicCmd(o),
		newBookingsCreateCmd(o),
		newBookingsUpdateCmd(o),
		newBookingsCancelCmd(o),
	)
	return cmd
}

// parseIntSlice converts repeated --action-code values into the []int the API wants.
func parseIntSlice(flag string, values []string) ([]int, error) {
	if len(values) == 0 {
		return nil, nil
	}
	out := make([]int, 0, len(values))
	for _, v := range values {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("--%s must be an integer, got %q", flag, v)
		}
		out = append(out, n)
	}
	return out, nil
}

func newBookingsListCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List bookings",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			userID, _ := cmd.Flags().GetString("user-id")
			bookingTypeID, _ := cmd.Flags().GetInt("booking-type-id")
			personalNumber, _ := cmd.Flags().GetString("personal-number")
			arrivalStatus, _ := cmd.Flags().GetInt("arrival-status")
			reservedForOrganizationID, _ := cmd.Flags().GetInt("reserved-for-organization-id")
			reservedForDepartmentID, _ := cmd.Flags().GetInt("reserved-for-department-id")

			fromDate, _ := cmd.Flags().GetString("from-date")
			if fromDate, err = ValidateDate("from-date", fromDate); err != nil {
				return err
			}
			toDate, _ := cmd.Flags().GetString("to-date")
			if toDate, err = ValidateDate("to-date", toDate); err != nil {
				return err
			}

			q := url.Values{}
			SetBool01(cmd, q, "isBooked", "is-booked")
			SetStr(q, "userId", userID)
			// clinicId is an optional filter here, so read o.ClinicID directly
			// instead of o.Clinic(), which errors when it is unset.
			SetStr(q, "clinicId", o.ClinicID)
			SetInt(q, "bookingTypeId", bookingTypeID)
			SetStr(q, "fromDate", fromDate)
			SetStr(q, "toDate", toDate)
			SetInt(q, "reservedForOrganizationId", reservedForOrganizationID)
			SetInt(q, "reservedForDepartmentId", reservedForDepartmentID)
			SetStr(q, "personalNumber", personalNumber)
			if cmd.Flags().Changed("arrival-status") {
				q.Set("arrivalStatus", strconv.Itoa(arrivalStatus))
			}
			SetPagination(cmd, q)

			bookings, err := api.ListBookings(client, q)
			if err != nil {
				return ScopeHint(err, "bookings:read")
			}

			return EmitEmpty(o, bookings, len(bookings) == 0, "No bookings found.", func() {
				for _, b := range bookings {
					fmt.Printf("%s  %s %s-%s  %s\n", b.ID, b.Date, b.StartTime, b.EndTime, b.BookingType.Name)
					if b.Patient.FirstName != "" || b.Patient.LastName != "" {
						fmt.Printf("  └─ patient %s %s (%s)\n", b.Patient.FirstName, b.Patient.LastName, b.Patient.PersonalNumber)
					}
					if b.User.FirstName != "" || b.User.LastName != "" {
						fmt.Printf("  └─ user %s %s %s\n", b.User.Title, b.User.FirstName, b.User.LastName)
					}
				}
			})
		},
	}

	cmd.Flags().Bool("is-booked", false, "Filter on booked (1) or unbooked (0) slots")
	cmd.Flags().String("user-id", "", "Filter on caregiver user ID")
	cmd.Flags().Int("booking-type-id", 0, "Filter on booking type ID")
	cmd.Flags().String("from-date", "", "Filter from date (YYYY-MM-DD)")
	cmd.Flags().String("to-date", "", "Filter to date (YYYY-MM-DD)")
	cmd.Flags().Int("reserved-for-organization-id", 0, "Filter on reserved organization ID")
	cmd.Flags().Int("reserved-for-department-id", 0, "Filter on reserved department ID")
	cmd.Flags().String("personal-number", "", "Filter on patient personal number")
	cmd.Flags().Int("arrival-status", 0, "Filter on arrival status")
	AddPagination(cmd)

	return cmd
}

func newBookingsListByClinicCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-by-clinic",
		Short: "List bookings for a clinic",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}
			clinicID, err := o.Clinic()
			if err != nil {
				return err
			}

			personalNumber, _ := cmd.Flags().GetString("personal-number")
			userID, _ := cmd.Flags().GetString("user-id")
			bookingTypeID, _ := cmd.Flags().GetInt("booking-type-id")
			reservedForOrganizationID, _ := cmd.Flags().GetInt("reserved-for-organization-id")
			reservedForDepartmentID, _ := cmd.Flags().GetInt("reserved-for-department-id")

			date, _ := cmd.Flags().GetString("date")
			if date, err = ValidateDate("date", date); err != nil {
				return err
			}
			fromDate, _ := cmd.Flags().GetString("from-date")
			if fromDate, err = ValidateDate("from-date", fromDate); err != nil {
				return err
			}
			toDate, _ := cmd.Flags().GetString("to-date")
			if toDate, err = ValidateDate("to-date", toDate); err != nil {
				return err
			}

			q := url.Values{}
			SetStr(q, "personalNumber", personalNumber)
			SetStr(q, "date", date)
			SetStr(q, "fromDate", fromDate)
			SetStr(q, "toDate", toDate)
			SetBool01(cmd, q, "arrived", "arrived")
			SetBool(cmd, q, "selfService", "self-service")
			SetInt(q, "bookingTypeId", bookingTypeID)
			SetBool(cmd, q, "booked", "booked")
			SetStr(q, "userId", userID)
			SetInt(q, "reservedForOrganizationId", reservedForOrganizationID)
			SetInt(q, "reservedForDepartmentId", reservedForDepartmentID)
			SetPagination(cmd, q)

			bookings, err := api.ListClinicBookings(client, clinicID, q)
			if err != nil {
				return err
			}

			return EmitEmpty(o, bookings, len(bookings) == 0, "No bookings found.", func() {
				for _, b := range bookings {
					fmt.Printf("%s  %s %s-%s  %s\n", b.ID, b.Date, b.StartTime, b.EndTime, b.BookingType)
					fmt.Printf("  └─ patient %s %s (%s)  %s\n", b.Patient.FirstName, b.Patient.LastName, b.Patient.PersonalNumber, b.BookedPatientType)
					fmt.Printf("  └─ caregiver %s %s %s\n", b.Caregiver.Title, b.Caregiver.FirstName, b.Caregiver.LastName)
				}
			})
		},
	}

	cmd.Flags().String("personal-number", "", "Filter on patient personal number")
	cmd.Flags().String("date", "", "Filter on a single date (YYYY-MM-DD)")
	cmd.Flags().String("from-date", "", "Filter from date (YYYY-MM-DD)")
	cmd.Flags().String("to-date", "", "Filter to date (YYYY-MM-DD)")
	cmd.Flags().Bool("arrived", false, "Filter on arrived (1) or not arrived (0)")
	cmd.Flags().Bool("self-service", false, "Filter on self-service bookings")
	cmd.Flags().Int("booking-type-id", 0, "Filter on booking type ID")
	cmd.Flags().Bool("booked", false, "Filter on booked slots")
	cmd.Flags().String("user-id", "", "Filter on caregiver user ID")
	cmd.Flags().Int("reserved-for-organization-id", 0, "Filter on reserved organization ID")
	cmd.Flags().Int("reserved-for-department-id", 0, "Filter on reserved department ID")
	AddPagination(cmd)

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
			clinicID, err := o.Clinic()
			if err != nil {
				return err
			}

			userID, _ := cmd.Flags().GetString("user-id")
			patientID, _ := cmd.Flags().GetString("patient-id")
			bookingTypeID, _ := cmd.Flags().GetInt("booking-type-id")
			patientTypeID, _ := cmd.Flags().GetInt("patient-type-id")

			date, _ := cmd.Flags().GetString("date")
			if date, err = ValidateDate("date", date); err != nil {
				return err
			}
			startTime, _ := cmd.Flags().GetString("start-time")
			if startTime, err = ValidateTime("start-time", startTime); err != nil {
				return err
			}
			endTime, _ := cmd.Flags().GetString("end-time")
			if endTime, err = ValidateTime("end-time", endTime); err != nil {
				return err
			}

			injuryNumber, _ := cmd.Flags().GetString("injury-number")
			note, _ := cmd.Flags().GetString("note")
			previousVisitID, _ := cmd.Flags().GetString("previous-visit-id")
			reservedForOrganizationID, _ := cmd.Flags().GetInt("reserved-for-organization-id")
			reservedForDepartmentID, _ := cmd.Flags().GetInt("reserved-for-department-id")

			rawActionCodes, _ := cmd.Flags().GetStringSlice("action-code")
			actionCodes, err := parseIntSlice("action-code", rawActionCodes)
			if err != nil {
				return err
			}

			req := api.CreateBookingRequest{
				UserID:                    userID,
				ClinicID:                  clinicID,
				PatientID:                 patientID,
				BookingTypeID:             bookingTypeID,
				PatientTypeID:             patientTypeID,
				Date:                      date,
				StartTime:                 startTime,
				EndTime:                   endTime,
				InjuryNumber:              injuryNumber,
				Note:                      note,
				PreviousVisitID:           previousVisitID,
				ActionCodes:               actionCodes,
				ReservedForOrganizationID: reservedForOrganizationID,
				ReservedForDepartmentID:   reservedForDepartmentID,
			}

			// false and 0 are meaningful values here, so only send them when
			// the user actually passed the flag.
			if cmd.Flags().Changed("fee") {
				fee, _ := cmd.Flags().GetFloat64("fee")
				req.Fee = &fee
			}
			if cmd.Flags().Changed("sms-reminder") {
				v, _ := cmd.Flags().GetBool("sms-reminder")
				req.SMSReminder = &v
			}
			if cmd.Flags().Changed("sms-confirmation") {
				v, _ := cmd.Flags().GetBool("sms-confirmation")
				req.SMSConfirmation = &v
			}
			if cmd.Flags().Changed("allow-overlap") {
				v, _ := cmd.Flags().GetBool("allow-overlap")
				req.AllowOverlap = &v
			}

			booking, err := api.CreateBooking(client, req)
			if err != nil {
				return ScopeHint(err, "bookings:write")
			}

			return Emit(o, booking, func() {
				fmt.Printf("Booking created: %s\n", booking.ID)
				fmt.Printf("  %s %s-%s\n", booking.Date, booking.StartTime, booking.EndTime)
				fmt.Printf("  Patient: %s %s\n", booking.Patient.FirstName, booking.Patient.LastName)
				fmt.Printf("  Caregiver: %s %s\n", booking.Caregiver.FirstName, booking.Caregiver.LastName)
				fmt.Printf("  Booking type: %s\n", booking.BookingType.Name)
			})
		},
	}

	cmd.Flags().String("user-id", "", "Caregiver user ID (required)")
	cmd.Flags().String("patient-id", "", "Patient ID (required)")
	cmd.Flags().Int("booking-type-id", 0, "Booking type ID (required)")
	cmd.Flags().Int("patient-type-id", 0, "Patient type ID (required)")
	cmd.Flags().String("date", "", "Booking date, YYYY-MM-DD (required)")
	cmd.Flags().String("start-time", "", "Start time, HH:MM (required)")
	cmd.Flags().String("end-time", "", "End time, HH:MM (required)")
	cmd.Flags().String("injury-number", "", "Injury number")
	cmd.Flags().String("note", "", "Booking note")
	cmd.Flags().String("previous-visit-id", "", "Previous visit ID")
	cmd.Flags().Float64("fee", 0, "Patient fee")
	cmd.Flags().Bool("sms-reminder", false, "Send an SMS reminder")
	cmd.Flags().Bool("sms-confirmation", false, "Send an SMS confirmation")
	cmd.Flags().StringSlice("action-code", nil, "Action code ID (repeatable)")
	cmd.Flags().Bool("allow-overlap", false, "Allow the booking to overlap an existing one")
	cmd.Flags().Int("reserved-for-organization-id", 0, "Reserve for organization ID")
	cmd.Flags().Int("reserved-for-department-id", 0, "Reserve for department ID")
	cmd.MarkFlagRequired("user-id")
	cmd.MarkFlagRequired("patient-id")
	cmd.MarkFlagRequired("booking-type-id")
	cmd.MarkFlagRequired("patient-type-id")
	cmd.MarkFlagRequired("date")
	cmd.MarkFlagRequired("start-time")
	cmd.MarkFlagRequired("end-time")

	return cmd
}

func newBookingsUpdateCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <bookingId>",
		Short: "Update a booking",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			changed := false
			for _, name := range []string{
				"requested-by-user-id", "booking-type-id", "caregiver-id",
				"reserved-for-organization-id", "reserved-for-department-id",
				"patient-id", "date", "start-time", "end-time", "note",
				"injury-number", "action-code",
			} {
				if cmd.Flags().Changed(name) {
					changed = true
					break
				}
			}
			if !changed {
				return fmt.Errorf("nothing to update: pass at least one field flag")
			}

			date, _ := cmd.Flags().GetString("date")
			if date, err = ValidateDate("date", date); err != nil {
				return err
			}
			startTime, _ := cmd.Flags().GetString("start-time")
			if startTime, err = ValidateTime("start-time", startTime); err != nil {
				return err
			}
			endTime, _ := cmd.Flags().GetString("end-time")
			if endTime, err = ValidateTime("end-time", endTime); err != nil {
				return err
			}

			requestedByUserID, _ := cmd.Flags().GetString("requested-by-user-id")
			bookingTypeID, _ := cmd.Flags().GetInt("booking-type-id")
			caregiverID, _ := cmd.Flags().GetString("caregiver-id")
			reservedForOrganizationID, _ := cmd.Flags().GetInt("reserved-for-organization-id")
			reservedForDepartmentID, _ := cmd.Flags().GetInt("reserved-for-department-id")
			patientID, _ := cmd.Flags().GetString("patient-id")
			note, _ := cmd.Flags().GetString("note")
			injuryNumber, _ := cmd.Flags().GetString("injury-number")

			rawActionCodes, _ := cmd.Flags().GetStringSlice("action-code")
			actionCodes, err := parseIntSlice("action-code", rawActionCodes)
			if err != nil {
				return err
			}

			req := api.UpdateBookingRequest{
				RequestedByUserID:         requestedByUserID,
				BookingTypeID:             bookingTypeID,
				ReservedForOrganizationID: reservedForOrganizationID,
				ReservedForDepartmentID:   reservedForDepartmentID,
				PatientID:                 patientID,
				Date:                      date,
				StartTime:                 startTime,
				EndTime:                   endTime,
				Note:                      note,
				InjuryNumber:              injuryNumber,
				ActionCodes:               actionCodes,
			}
			if caregiverID != "" {
				req.Caregiver = &api.IDRef{ID: caregiverID}
			}

			booking, err := api.UpdateBooking(client, args[0], req)
			if err != nil {
				return ScopeHint(err, "bookings:write")
			}

			return Emit(o, booking, func() {
				fmt.Printf("Booking updated: %s\n", booking.ID)
				fmt.Printf("  %s %s-%s\n", booking.Date, booking.StartTime, booking.EndTime)
				fmt.Printf("  Patient: %s %s\n", booking.Patient.FirstName, booking.Patient.LastName)
				fmt.Printf("  Caregiver: %s %s\n", booking.Caregiver.FirstName, booking.Caregiver.LastName)
				fmt.Printf("  Booking type: %s\n", booking.BookingType)
			})
		},
	}

	cmd.Flags().String("requested-by-user-id", "", "User ID requesting the change")
	cmd.Flags().Int("booking-type-id", 0, "New booking type ID")
	cmd.Flags().String("caregiver-id", "", "New caregiver user ID")
	cmd.Flags().Int("reserved-for-organization-id", 0, "Reserve for organization ID")
	cmd.Flags().Int("reserved-for-department-id", 0, "Reserve for department ID")
	cmd.Flags().String("patient-id", "", "New patient ID")
	cmd.Flags().String("date", "", "New date, YYYY-MM-DD")
	cmd.Flags().String("start-time", "", "New start time, HH:MM")
	cmd.Flags().String("end-time", "", "New end time, HH:MM")
	cmd.Flags().String("note", "", "New booking note")
	cmd.Flags().String("injury-number", "", "New injury number")
	cmd.Flags().StringSlice("action-code", nil, "Action code ID (repeatable, replaces the set)")

	return cmd
}

func newBookingsCancelCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel <bookingId>",
		Short: "Cancel a booking",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			userID, _ := cmd.Flags().GetString("user-id")

			if err := api.CancelBooking(client, args[0], api.DeleteBookingRequest{UserID: userID}); err != nil {
				return ScopeHint(err, "bookings:write")
			}

			result := map[string]string{"id": args[0], "status": "cancelled"}
			return Emit(o, result, func() {
				fmt.Printf("Booking %s cancelled\n", args[0])
			})
		},
	}

	cmd.Flags().String("user-id", "", "User ID performing the cancellation (required)")
	cmd.MarkFlagRequired("user-id")

	return cmd
}
