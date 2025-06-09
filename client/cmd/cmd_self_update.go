package cmd

import (
	"fmt"
	"wowa/core"
	"wowa/spinny"

	"github.com/spf13/cobra"
)

func SetupSelfUpdateCmd(rootCmd *cobra.Command, selfUpdateManager *core.SelfUpdateManager) {
	var selfUpdateCmd = &cobra.Command{
		Use:     "self-update",
		Aliases: []string{"su"},
		Short:   "Check and download new wowa updates",
		RunE: func(cmd *cobra.Command, args []string) error {
			var spinners = spinny.NewManager()
			spinners.Start()
			defer spinners.Stop()

			var spinner = spinners.NewSpinner("checking for updates")

			result, err := selfUpdateManager.UpdateToLatest()
			if err != nil {
				spinner.Fail(err.Error())
				return err
			}

			if result.Updated {
				spinner.Succeed(fmt.Sprintf("updated wowa to %s", result.ToVersion))
			} else {
				spinner.Info("wowa is already up to date")
			}

			return nil
		},
	}
	rootCmd.AddCommand(selfUpdateCmd)
}
