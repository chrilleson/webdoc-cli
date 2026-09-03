package commands

import (
	"fmt"
	"net/url"

	"github.com/chrilleson/webdoc-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewVisitsCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "visits",
		Short: "Manage visits",
	}
	cmd.AddCommand(
		newVisitsListCmd(o),
		newVisitsCreateCmd(o),
		newVisitsSignCmd(o),
	)
	return cmd
}

func newVisitsListCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List visits for a clinic",
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
			recordTemplateID, _ := cmd.Flags().GetString("record-template-id")

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
			SetStr(q, "userId", userID)
			SetStr(q, "recordTemplateId", recordTemplateID)
			SetBool01(cmd, q, "isSigned", "is-signed")

			visits, err := api.ListClinicVisits(client, clinicID, q)
			if err != nil {
				return err
			}

			return EmitEmpty(o, visits, len(visits) == 0, "No visits found.", func() {
				for _, v := range visits {
					fmt.Printf("%s  %s  %s %s\n", v.ID, v.VisitDate, v.Patient.FirstName, v.Patient.LastName)
					fmt.Printf("  └─ caregiver %s %s %s\n", v.Caregiver.Title, v.Caregiver.FirstName, v.Caregiver.LastName)
				}
			})
		},
	}

	cmd.Flags().String("personal-number", "", "Filter on patient personal number")
	cmd.Flags().String("date", "", "Filter on a single date (YYYY-MM-DD)")
	cmd.Flags().String("from-date", "", "Filter from date (YYYY-MM-DD)")
	cmd.Flags().String("to-date", "", "Filter to date (YYYY-MM-DD)")
	cmd.Flags().String("user-id", "", "Filter on caregiver user ID")
	cmd.Flags().String("record-template-id", "", "Filter on record template ID")
	cmd.Flags().Bool("is-signed", false, "Filter on signed (1) or unsigned (0) visits")

	return cmd
}

func newVisitsCreateCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a visit from a booking",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			bookingID, _ := cmd.Flags().GetString("booking-id")
			visitDate, _ := cmd.Flags().GetString("visit-date")
			if visitDate, err = ValidateDateTime("visit-date", visitDate); err != nil {
				return err
			}

			visit, err := api.CreateVisit(client, api.CreateVisitRequest{
				BookingID: bookingID,
				VisitDate: visitDate,
			})
			if err != nil {
				return err
			}

			return Emit(o, visit, func() {
				fmt.Printf("Visit created: %s\n", visit.ID)
				fmt.Printf("  Visit date: %s\n", visit.VisitDate)
				fmt.Printf("  Booking: %s\n", visit.BookingID)
				fmt.Printf("  Patient: %s %s (%s)\n", visit.Patient.FirstName, visit.Patient.LastName, visit.Patient.PersonalNumber)
			})
		},
	}

	cmd.Flags().String("booking-id", "", "Booking to create the visit from (required)")
	cmd.Flags().String("visit-date", "", "Visit date, YYYY-MM-DD HH:MM:SS (required)")
	cmd.MarkFlagRequired("booking-id")
	cmd.MarkFlagRequired("visit-date")

	return cmd
}

func newVisitsSignCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign <visitId>",
		Short: "Sign a visit's record",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			userID, _ := cmd.Flags().GetString("user-id")

			if err := api.SignVisit(client, args[0], api.SignVisitRequest{UserID: userID}); err != nil {
				return ScopeHint(err, "record-signatures:write")
			}

			result := map[string]string{"id": args[0], "status": "signed"}
			return Emit(o, result, func() {
				fmt.Printf("Visit %s signed\n", args[0])
			})
		},
	}

	cmd.Flags().String("user-id", "", "User ID signing the record (required)")
	cmd.MarkFlagRequired("user-id")

	return cmd
}
