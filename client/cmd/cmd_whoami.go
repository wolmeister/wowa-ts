package cmd

import (
	"fmt"
	"wowa/core"
	"wowa/utils"

	"github.com/spf13/cobra"
)

func SetupWhoamiCmd(rootCmd *cobra.Command, userManager *core.UserManager) {
	var addCmd = &cobra.Command{
		Use:   "whoami",
		Short: "Display the user email currently logged in",
		RunE: func(cmd *cobra.Command, args []string) error {
			email, err := userManager.GetUserEmail()
			if err != nil {
				return err
			}

			if email == "" {
				fmt.Println(utils.AnsiYellow, "null", utils.AnsiReset)
				return nil
			}

			fmt.Println(utils.AnsiBlue, email, utils.AnsiReset)
			return nil
		},
	}

	rootCmd.AddCommand(addCmd)
}
