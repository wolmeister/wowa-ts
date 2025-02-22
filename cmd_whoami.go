package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

func SetupWhoamiCmd(rootCmd *cobra.Command, userManager *UserManager) {
	var addCmd = &cobra.Command{
		Use:   "whoami",
		Short: "Display the user email currently logged in",
		RunE: func(cmd *cobra.Command, args []string) error {
			email, err := userManager.GetUserEmail()
			if err != nil {
				return err
			}

			if email == "" {
				fmt.Println(AnsiYellow, "null", AnsiReset)
				return nil
			}

			fmt.Println(AnsiBlue, email, AnsiReset)
			return nil
		},
	}

	rootCmd.AddCommand(addCmd)
}
