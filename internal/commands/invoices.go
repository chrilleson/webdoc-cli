package commands

import (
	"fmt"
	"net/url"

	"github.com/chrilleson/webdoc-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewInvoicesCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invoices",
		Short: "Manage clinic invoices",
	}
	cmd.AddCommand(
		newInvoicesListCmd(o),
		newInvoicesListCreditCmd(o),
		newInvoicesCreateCmd(o),
		newInvoicesUpdateCmd(o),
	)
	return cmd
}

// invoicesAddFilterFlags registers the filters shared by list and list-credit.
func invoicesAddFilterFlags(cmd *cobra.Command) {
	cmd.Flags().String("personal-number", "", "Filter by patient personal number")
	cmd.Flags().String("date", "", "Filter by date (YYYY-MM-DD)")
	cmd.Flags().String("from-date", "", "Filter from date (YYYY-MM-DD)")
	cmd.Flags().String("to-date", "", "Filter to date (YYYY-MM-DD)")
	cmd.Flags().Bool("is-exported", false, "Filter on exported invoices")
}

func invoicesFilterQuery(cmd *cobra.Command) (url.Values, error) {
	personalNumber, _ := cmd.Flags().GetString("personal-number")
	rawDate, _ := cmd.Flags().GetString("date")
	rawFrom, _ := cmd.Flags().GetString("from-date")
	rawTo, _ := cmd.Flags().GetString("to-date")

	date, err := ValidateDate("date", rawDate)
	if err != nil {
		return nil, err
	}
	fromDate, err := ValidateDate("from-date", rawFrom)
	if err != nil {
		return nil, err
	}
	toDate, err := ValidateDate("to-date", rawTo)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	SetStr(q, "personalNumber", personalNumber)
	SetStr(q, "date", date)
	SetStr(q, "fromDate", fromDate)
	SetStr(q, "toDate", toDate)
	SetBool01(cmd, q, "isExported", "is-exported")
	return q, nil
}

func invoicesPrint(invoices []api.Invoice) {
	for _, inv := range invoices {
		fmt.Printf("%-12s  %-10s  %10s %s  %s %s\n",
			inv.InvoiceNumber, inv.InvoiceStatus, inv.TotalAmount, inv.Currency,
			inv.Patient.FirstName, inv.Patient.LastName)
		fmt.Printf("  └─ created %s  expires %s\n", inv.CreatedAt, inv.ExpiryDate)
	}
}

func newInvoicesListCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List invoices for a clinic",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}
			clinicID, err := o.Clinic()
			if err != nil {
				return err
			}

			q, err := invoicesFilterQuery(cmd)
			if err != nil {
				return err
			}
			status, _ := cmd.Flags().GetInt("status")
			SetInt(q, "status", status)

			invoices, err := api.ListInvoices(client, clinicID, q)
			if err != nil {
				return err
			}

			return EmitEmpty(o, invoices, len(invoices) == 0, "No invoices found.", func() {
				invoicesPrint(invoices)
			})
		},
	}
	invoicesAddFilterFlags(cmd)
	cmd.Flags().Int("status", 0, "Filter by invoice status")

	return cmd
}

func newInvoicesListCreditCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-credit",
		Short: "List credit invoices for a clinic",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}
			clinicID, err := o.Clinic()
			if err != nil {
				return err
			}

			q, err := invoicesFilterQuery(cmd)
			if err != nil {
				return err
			}

			invoices, err := api.ListCreditInvoices(client, clinicID, q)
			if err != nil {
				return err
			}

			return EmitEmpty(o, invoices, len(invoices) == 0, "No invoices found.", func() {
				invoicesPrint(invoices)
			})
		},
	}
	invoicesAddFilterFlags(cmd)

	return cmd
}

func newInvoicesCreateCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an invoice from a visit",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}
			clinicID, err := o.Clinic()
			if err != nil {
				return err
			}

			visitID, _ := cmd.Flags().GetString("visit-id")
			totalPayment, _ := cmd.Flags().GetString("total-payment")

			invoice, err := api.CreateInvoice(client, clinicID, api.CreateInvoiceRequest{
				VisitID:      visitID,
				TotalPayment: totalPayment,
			})
			if err != nil {
				return err
			}

			return Emit(o, invoice, func() {
				fmt.Printf("Invoice created: %s (%s)\n", invoice.InvoiceNumber, invoice.InvoiceID)
				fmt.Printf("  Total: %s %s\n", invoice.TotalAmount, invoice.Currency)
				fmt.Printf("  Expires: %s\n", invoice.ExpiryDate)
			})
		},
	}
	cmd.Flags().String("visit-id", "", "Visit ID (required)")
	cmd.Flags().String("total-payment", "", "Total payment amount, e.g. 123 (required)")
	cmd.MarkFlagRequired("visit-id")
	cmd.MarkFlagRequired("total-payment")

	return cmd
}

func newInvoicesUpdateCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <invoiceId>",
		Short: "Update an invoice",
		Long: "Update an invoice's total payment or payment method.\n\n" +
			"Payment methods: 1 = Card, 2 = Cash, 3 = Exported invoice, 4 = Paid invoice.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			totalPayment, _ := cmd.Flags().GetString("total-payment")
			paymentMethod, _ := cmd.Flags().GetString("payment-method")
			if totalPayment == "" && paymentMethod == "" {
				return fmt.Errorf("nothing to update: set --total-payment and/or --payment-method")
			}

			client, err := o.Client()
			if err != nil {
				return err
			}
			clinicID, err := o.Clinic()
			if err != nil {
				return err
			}

			invoice, err := api.UpdateInvoice(client, clinicID, args[0], api.UpdateInvoiceRequest{
				TotalPayment:  totalPayment,
				PaymentMethod: paymentMethod,
			})
			if err != nil {
				return err
			}

			return Emit(o, invoice, func() {
				fmt.Printf("Invoice %s updated\n", invoice.InvoiceNumber)
				fmt.Printf("  Status: %s\n", invoice.InvoiceStatus)
				fmt.Printf("  Total: %s %s\n", invoice.TotalAmount, invoice.Currency)
			})
		},
	}
	cmd.Flags().String("total-payment", "", "New total payment amount, e.g. 123")
	cmd.Flags().String("payment-method", "", "Payment method: 1=Card, 2=Cash, 3=Exported invoice, 4=Paid invoice")

	return cmd
}
