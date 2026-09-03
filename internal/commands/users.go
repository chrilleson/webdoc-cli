package commands

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/chrilleson/webdoc-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewUsersCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users",
		Short: "Manage users",
	}
	cmd.AddCommand(
		newUsersSearchCmd(o),
		newUsersListCmd(o),
		newUsersGetCmd(o),
		newUsersPermissionsCmd(o),
		newUsersListByClinicCmd(o),
	)
	return cmd
}

func newUsersSearchCmd(o *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "search <personalNumber>",
		Short: "Search for a user by personalNumber",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			users, err := api.SearchUsers(client, args[0])
			if err != nil {
				return err
			}

			return EmitEmpty(o, users, len(users) == 0, "No users found", func() {
				usersPrintLines(users)
			})
		},
	}
}

func usersPrintLines(users []api.User) {
	for _, u := range users {
		fmt.Printf("%s %s %s (%s)\n", u.ID, u.FirstName, u.LastName, u.PersonalNumber)
		for _, c := range u.Clinics {
			fmt.Printf("  └─ %s  %s\n", c.ID, c.Name)
		}
	}
}

func newUsersListCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List users, optionally filtered",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			firstName, _ := cmd.Flags().GetString("first-name")
			lastName, _ := cmd.Flags().GetString("last-name")
			personalNumber, _ := cmd.Flags().GetString("personal-number")
			hsaID, _ := cmd.Flags().GetString("hsa-id")
			id, _ := cmd.Flags().GetString("id")

			q := url.Values{}
			SetStr(q, "firstName", firstName)
			SetStr(q, "lastName", lastName)
			SetStr(q, "personalNumber", personalNumber)
			SetStr(q, "hsaId", hsaID)
			SetStr(q, "id", id)

			users, err := api.ListUsers(client, q)
			if err != nil {
				return ScopeHint(err, "users:read")
			}

			return EmitEmpty(o, users, len(users) == 0, "No users found.", func() {
				usersPrintLines(users)
			})
		},
	}
	cmd.Flags().String("first-name", "", "Filter by first name")
	cmd.Flags().String("last-name", "", "Filter by last name")
	cmd.Flags().String("personal-number", "", "Filter by personal number")
	cmd.Flags().String("hsa-id", "", "Filter by HSA ID")
	cmd.Flags().String("id", "", "Filter by user ID")

	return cmd
}

func newUsersGetCmd(o *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <userId>",
		Short: "Get a user by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			u, err := api.GetUser(client, args[0])
			if err != nil {
				return ScopeHint(err, "users:read")
			}

			return Emit(o, u, func() {
				fmt.Printf("%s  %s %s  (%s)\n", u.ID, u.FirstName, u.LastName, u.PersonalNumber)
				if u.HsaID != "" {
					fmt.Printf("  HSA ID: %s\n", u.HsaID)
				}
				if u.Email != "" {
					fmt.Printf("  Email: %s\n", u.Email)
				}
				if phone := u.Phone(); phone != "" {
					fmt.Printf("  Phone: %s\n", phone)
				}
				if u.LastLoggedIn != "" {
					fmt.Printf("  Last logged in: %s\n", u.LastLoggedIn)
				}
				for _, c := range u.Clinics {
					fmt.Printf("  └─ %s  %s\n", c.ID, c.Name)
				}
				fmt.Println("  Settings:")
				if cc := u.Settings.CostCentre(); cc != "" {
					fmt.Printf("    Default cost centre: %s\n", cc)
				}
				if u.Settings.DefaultRecordTemplateID != nil {
					fmt.Printf("    Default record template: %s\n", *u.Settings.DefaultRecordTemplateID)
				}
				if u.Settings.DefaultPatientTypeID != "" {
					fmt.Printf("    Default patient type: %s\n", u.Settings.DefaultPatientTypeID)
				}
			})
		},
	}
}

// usersPermissionIsSet reports whether a permission value means the user actually
// has it. Webdoc returns a mix of "1"/"0", "yes"/"no" and
// "full"/"read"/"read/write"/"none" across permission names.
func usersPermissionIsSet(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "0", "no", "none", "false":
		return false
	}
	return true
}

func newUsersPermissionsCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "permissions <uuid>",
		Short: "Show a user's permissions per clinic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			perms, err := api.GetUserPermissions(client, args[0])
			if err != nil {
				return ScopeHint(err, "users:read")
			}

			all, _ := cmd.Flags().GetBool("all")

			return EmitEmpty(o, perms, len(perms) == 0, "No permissions found.", func() {
				clinicIDs := make([]string, 0, len(perms))
				for clinicID := range perms {
					clinicIDs = append(clinicIDs, clinicID)
				}
				sort.Strings(clinicIDs)

				for _, clinicID := range clinicIDs {
					fmt.Println(clinicID)
					names := make([]string, 0, len(perms[clinicID]))
					for name := range perms[clinicID] {
						names = append(names, name)
					}
					sort.Strings(names)
					for _, name := range names {
						value := perms[clinicID][name]
						if !all && !usersPermissionIsSet(value) {
							continue
						}
						fmt.Printf("  %-40s %s\n", name, value)
					}
				}
			})
		},
	}
	cmd.Flags().Bool("all", false, "Include permissions the user does not have")

	return cmd
}

func newUsersListByClinicCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-by-clinic",
		Short: "List the users of a clinic",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}
			clinicID, err := o.Clinic()
			if err != nil {
				return err
			}

			q := url.Values{}
			SetBool(cmd, q, "isCaregiver", "is-caregiver")

			users, err := api.ListClinicUsers(client, clinicID, q)
			if err != nil {
				return ScopeHint(err, "users:read")
			}

			return EmitEmpty(o, users, len(users) == 0,
				fmt.Sprintf("No users found for clinic %s.", clinicID), func() {
					for _, u := range users {
						fmt.Printf("%s  %s %s\n", u.ID, u.FirstName, u.LastName)
					}
				})
		},
	}
	cmd.Flags().Bool("is-caregiver", false, "Only users that are (or are not) caregivers")

	return cmd
}
