package commands

import (
	"fmt"

	"github.com/chrilleson/webdoc-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewPatientsCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "patients",
		Short: "Manage patients",
	}
	cmd.AddCommand(
		newPatientsSearchCmd(o),
		newPatientsListCmd(o),
		newPatientsCreateCmd(o),
		newPatientsCreateWithoutSSNCmd(o),
		newPatientsUpdateCmd(o),
		newPatientsNationalRegistryCmd(o),
		newPatientsSetCurrentCmd(o),
	)
	return cmd
}

func newPatientsSearchCmd(o *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "search <personalNumber>",
		Short: "Search for a patient by personal number",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			patients, err := api.GetPatients(client, args[0])
			if err != nil {
				return err
			}

			return EmitEmpty(o, patients, len(patients) == 0, "No patients found.", func() {
				for _, p := range patients {
					fmt.Printf("%s  %s %s  (%s)\n", p.ID, p.FirstName, p.LastName, p.PersonalNumber)
					fmt.Printf("  Born: %s  Gender: %s  Nationality: %s\n", p.BirthDate, p.Gender, p.Nationality)
					fmt.Printf("  Address: %s, %s %s\n", p.Address.StreetName, p.Address.ZipCode, p.Address.City)
					if p.Email != "" {
						fmt.Printf("  Email: %s\n", p.Email)
					}
					if p.MobilePhoneNumber != "" {
						fmt.Printf("  Mobile: %s\n", p.MobilePhoneNumber)
					}
					fmt.Printf("  Patient type: %s (%s)\n", p.PatientType.Name, p.PatientType.Type)
					if p.Organization != nil {
						fmt.Printf("  Organization: %s\n", p.Organization.Name)
					}
				}
			})
		},
	}
}

func newPatientsCreateCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new patient",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			personalNumber, _ := cmd.Flags().GetString("personal-number")
			clinicName, _ := cmd.Flags().GetString("clinic-name")
			clinicHsaID, _ := cmd.Flags().GetString("clinic-hsa-id")
			countyName, _ := cmd.Flags().GetString("county-name")
			patientType, _ := cmd.Flags().GetString("patient-type")
			email, _ := cmd.Flags().GetString("email")
			mobile, _ := cmd.Flags().GetString("mobile")

			req := api.CreatePatientRequest{
				PersonalNumber: personalNumber,
				ListInfo: api.CreatePatientListInfo{
					Name:       clinicName,
					HsaID:      clinicHsaID,
					CountyName: countyName,
				},
				PatientType:  patientType,
				Email:        email,
				MobileNumber: mobile,
			}

			patient, err := api.CreatePatient(client, req)
			if err != nil {
				return err
			}

			return Emit(o, patient, func() {
				fmt.Printf("Patient created: %s\n", patient.ID)
				fmt.Printf("  %s %s  (%s)\n", patient.FirstName, patient.LastName, patient.PersonalNumber)
				fmt.Printf("  Patient type: %s\n", patient.PatientType)
			})
		},
	}

	cmd.Flags().String("personal-number", "", "Personal number (required)")
	cmd.Flags().String("clinic-name", "", "Listed clinic name (required)")
	cmd.Flags().String("clinic-hsa-id", "", "Listed clinic HSA ID (required)")
	cmd.Flags().String("county-name", "", "County name (required)")
	cmd.Flags().String("patient-type", "", "Patient type name, e.g. Private (required)")
	cmd.Flags().String("email", "", "Email address")
	cmd.Flags().String("mobile", "", "Mobile number")
	cmd.MarkFlagRequired("personal-number")
	cmd.MarkFlagRequired("clinic-name")
	cmd.MarkFlagRequired("clinic-hsa-id")
	cmd.MarkFlagRequired("county-name")
	cmd.MarkFlagRequired("patient-type")

	return cmd
}

func newPatientsListCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Fetch a patient via the deprecated GET /v1/patients (use `patients search`)",
		Long: "Fetch a single patient through the v1 endpoint.\n\n" +
			"The API documents this endpoint as deprecated in favour of GET /v2/patients,\n" +
			"which is what `patients search` uses. Despite the plural path it returns one\n" +
			"patient object, and the only filter sent is --personal-number.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			personalNumber, _ := cmd.Flags().GetString("personal-number")

			p, err := api.ListPatientsV1(client, personalNumber)
			if err != nil {
				return err
			}

			return Emit(o, p, func() {
				fmt.Printf("%s  %s %s  (%s)\n", p.ID, p.FirstName, p.LastName, p.PersonalNumber)
				fmt.Printf("  Born: %s  Gender: %s  Nationality: %s\n", p.BirthDate, p.Gender, p.Nationality)
				fmt.Printf("  Address: %s, %s %s\n", p.Address.StreetName, p.Address.ZipCode, p.Address.City)
				if p.Email != "" {
					fmt.Printf("  Email: %s\n", p.Email)
				}
				if p.MobilePhoneNumber != "" {
					fmt.Printf("  Mobile: %s\n", p.MobilePhoneNumber)
				}
				fmt.Printf("  Patient type: %s (%s)\n", p.PatientType.Name, p.PatientType.Type)
			})
		},
	}
	cmd.Flags().String("personal-number", "", "Personal number (required)")
	cmd.MarkFlagRequired("personal-number")

	return cmd
}

func newPatientsCreateWithoutSSNCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-without-ssn",
		Short: "Create a patient without a social security number",
		Long: "Create a patient with a reserve number (reservnummer/samordningsnummer)\n" +
			"or a foreign patient. The identification number is 6-8 digits followed by\n" +
			"4 digits or letters, optionally separated by a dash.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			identificationNumber, _ := cmd.Flags().GetString("identification-number")
			firstName, _ := cmd.Flags().GetString("first-name")
			lastName, _ := cmd.Flags().GetString("last-name")
			gender, _ := cmd.Flags().GetString("gender")
			patientTypeID, _ := cmd.Flags().GetInt("patient-type-id")
			clinicName, _ := cmd.Flags().GetString("clinic-name")
			clinicHsaID, _ := cmd.Flags().GetString("clinic-hsa-id")
			countyName, _ := cmd.Flags().GetString("county-name")
			streetName, _ := cmd.Flags().GetString("street-name")
			zipCode, _ := cmd.Flags().GetString("zip-code")
			city, _ := cmd.Flags().GetString("city")
			coAddress, _ := cmd.Flags().GetString("co-address")
			mobile, _ := cmd.Flags().GetString("mobile")
			email, _ := cmd.Flags().GetString("email")
			workPhone, _ := cmd.Flags().GetString("work-phone")
			homePhone, _ := cmd.Flags().GetString("home-phone")

			rawBirthDate, _ := cmd.Flags().GetString("birth-date")
			birthDate, err := ValidateDate("birth-date", rawBirthDate)
			if err != nil {
				return err
			}

			req := api.CreatePatientWithoutSSNRequest{
				IdentificationNumber: identificationNumber,
				FirstName:            firstName,
				LastName:             lastName,
				BirthDate:            birthDate,
				Gender:               gender,
				PatientTypeID:        patientTypeID,
				Contact: api.CreatedPatientContact{
					StreetName: streetName,
					ZipCode:    zipCode,
					City:       city,
					CoAddress:  coAddress,
					Mobile:     mobile,
					Email:      email,
					WorkPhone:  workPhone,
					HomePhone:  homePhone,
				},
				ListInfo: api.CreatePatientListInfo{
					Name:       clinicName,
					HsaID:      clinicHsaID,
					CountyName: countyName,
				},
			}

			patient, err := api.CreatePatientWithoutSSN(client, req)
			if err != nil {
				return ScopeHint(err, "patient:write")
			}

			return Emit(o, patient, func() {
				fmt.Printf("Patient created: %s\n", patient.ID)
				fmt.Printf("  %s %s  (%s)\n", patient.FirstName, patient.LastName, patient.IdentificationNumber)
				fmt.Printf("  Born: %s  Gender: %s\n", patient.BirthDate, patient.Gender)
				fmt.Printf("  Patient type: %s (%d)\n", patient.PatientType.Name, patient.PatientType.ID)
			})
		},
	}
	cmd.Flags().String("identification-number", "", "Identification number, e.g. 19121212-M123 (required)")
	cmd.Flags().String("first-name", "", "First name (required)")
	cmd.Flags().String("last-name", "", "Last name (required)")
	cmd.Flags().String("birth-date", "", "Birth date, YYYY-MM-DD (required)")
	cmd.Flags().String("gender", "", "Gender, M or K (required)")
	cmd.Flags().Int("patient-type-id", 0, "Patient type ID (required)")
	cmd.Flags().String("clinic-name", "", "Listed clinic name (required)")
	cmd.Flags().String("clinic-hsa-id", "", "Listed clinic HSA ID (required)")
	cmd.Flags().String("county-name", "", "County name (required)")
	cmd.Flags().String("street-name", "", "Street address")
	cmd.Flags().String("zip-code", "", "Zip code")
	cmd.Flags().String("city", "", "City")
	cmd.Flags().String("co-address", "", "C/O address")
	cmd.Flags().String("mobile", "", "Mobile number")
	cmd.Flags().String("email", "", "Email address")
	cmd.Flags().String("work-phone", "", "Work phone number")
	cmd.Flags().String("home-phone", "", "Home phone number")
	for _, name := range []string{
		"identification-number", "first-name", "last-name", "birth-date",
		"gender", "patient-type-id", "clinic-name", "clinic-hsa-id", "county-name",
	} {
		cmd.MarkFlagRequired(name)
	}

	return cmd
}

// patientsUpdateFlags are the field flags of `patients update`; at least one
// must be set or there is nothing to PATCH.
var patientsUpdateFlags = []string{
	"clinic-name", "clinic-hsa-id", "county-name", "patient-type", "email",
	"mobile", "first-name", "last-name", "spoken-name", "address", "co-address",
	"city", "zip-code", "state-code", "county-code",
}

func newPatientsUpdateCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <personalNumber>",
		Short: "Update an existing patient",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			changed := false
			for _, name := range patientsUpdateFlags {
				if cmd.Flags().Changed(name) {
					changed = true
					break
				}
			}
			if !changed {
				return fmt.Errorf("nothing to update: pass at least one field flag")
			}

			client, err := o.Client()
			if err != nil {
				return err
			}

			clinicName, _ := cmd.Flags().GetString("clinic-name")
			clinicHsaID, _ := cmd.Flags().GetString("clinic-hsa-id")
			countyName, _ := cmd.Flags().GetString("county-name")
			patientType, _ := cmd.Flags().GetString("patient-type")
			email, _ := cmd.Flags().GetString("email")
			mobile, _ := cmd.Flags().GetString("mobile")
			firstName, _ := cmd.Flags().GetString("first-name")
			lastName, _ := cmd.Flags().GetString("last-name")
			spokenName, _ := cmd.Flags().GetString("spoken-name")
			address, _ := cmd.Flags().GetString("address")
			coAddress, _ := cmd.Flags().GetString("co-address")
			city, _ := cmd.Flags().GetString("city")
			zipCode, _ := cmd.Flags().GetString("zip-code")
			stateCode, _ := cmd.Flags().GetString("state-code")
			countyCode, _ := cmd.Flags().GetString("county-code")

			req := api.UpdatePatientRequest{
				PatientType:  patientType,
				Email:        email,
				MobileNumber: mobile,
				FirstName:    firstName,
				LastName:     lastName,
				SpokenName:   spokenName,
				Address:      address,
				CoAddress:    coAddress,
				City:         city,
				ZipCode:      zipCode,
				StateCode:    stateCode,
				CountyCode:   countyCode,
			}
			if clinicName != "" || clinicHsaID != "" || countyName != "" {
				req.ListInfo = &api.CreatePatientListInfo{
					Name:       clinicName,
					HsaID:      clinicHsaID,
					CountyName: countyName,
				}
			}

			if err := api.UpdatePatient(client, args[0], req); err != nil {
				return ScopeHint(err, "patient:write")
			}

			return Emit(o, req, func() {
				fmt.Printf("Patient %s updated\n", args[0])
			})
		},
	}
	cmd.Flags().String("clinic-name", "", "Listed clinic name")
	cmd.Flags().String("clinic-hsa-id", "", "Listed clinic HSA ID")
	cmd.Flags().String("county-name", "", "Listed county name")
	cmd.Flags().String("patient-type", "", "Patient type name, e.g. Private")
	cmd.Flags().String("email", "", "Email address")
	cmd.Flags().String("mobile", "", "Mobile number")
	cmd.Flags().String("first-name", "", "First name")
	cmd.Flags().String("last-name", "", "Last name")
	cmd.Flags().String("spoken-name", "", "Spoken name")
	cmd.Flags().String("address", "", "Street address")
	cmd.Flags().String("co-address", "", "C/O address")
	cmd.Flags().String("city", "", "City")
	cmd.Flags().String("zip-code", "", "Zip code")
	cmd.Flags().String("state-code", "", "State code")
	cmd.Flags().String("county-code", "", "County code")

	return cmd
}

func newPatientsNationalRegistryCmd(o *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "national-registry <personalNumber>",
		Short: "Look a person up in the national registry (SPAR)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			p, err := api.GetNationalRegistryPerson(client, args[0])
			if err != nil {
				return ScopeHint(err, "patient:read")
			}

			return Emit(o, p, func() {
				fmt.Printf("%s %s\n", p.FirstName, p.LastName)
				fmt.Printf("  Address: %s, %s %s\n", p.Address, p.ZipCode, p.City)
				if p.CoAddress != "" {
					fmt.Printf("  C/O: %s\n", p.CoAddress)
				}
				if p.Parish != "" {
					fmt.Printf("  Parish: %s\n", p.Parish)
				}
				if p.CountyCode != "" {
					fmt.Printf("  County: %s\n", p.CountyCode)
				}
				if p.StateCode != "" {
					fmt.Printf("  State: %s\n", p.StateCode)
				}
				if p.Deceased != "" && p.Deceased != "0" {
					fmt.Printf("  Deceased: %s\n", p.Deceased)
				}
				if p.UnRegisterCauseCode != nil && *p.UnRegisterCauseCode != "" {
					fmt.Printf("  Unregistered: %s\n", *p.UnRegisterCauseCode)
				}
			})
		},
	}
}

func newPatientsSetCurrentCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-current <personalNumber>",
		Short: "Switch the current patient of the token's user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			showWarning, _ := cmd.Flags().GetBool("show-warning")

			req := api.SetCurrentPatientRequest{
				PersonalNumber: args[0],
				ShowWarning:    showWarning,
			}
			if err := api.SetCurrentPatient(client, req); err != nil {
				return ScopeHint(err, "patient:write")
			}

			return Emit(o, req, func() {
				fmt.Printf("Current patient set to %s\n", args[0])
			})
		},
	}
	cmd.Flags().Bool("show-warning", true, "Show the patient switch warning in Webdoc")

	return cmd
}
