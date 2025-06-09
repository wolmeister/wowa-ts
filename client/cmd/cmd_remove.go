package cmd

import (
	"fmt"
	"wowa/core"
	"wowa/spinny"

	"github.com/spf13/cobra"
)

func SetupRemoveCmd(rootCmd *cobra.Command, addonManager *core.AddonManager) {
	var removeCmd = &cobra.Command{
		Use:     "rm <id>",
		Aliases: []string{"remove"},
		Short:   "Uninstall an addon",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			var gameVersion core.GameVersion
			if cmd.Flag("retail").Value.String() == "true" {
				gameVersion = core.Retail
			} else {
				gameVersion = core.Classic
			}

			var spinners = spinny.NewManager()
			spinners.Start()
			defer spinners.Stop()

			var spinner = spinners.NewSpinner(fmt.Sprintf("Removing %s (%s)", id, gameVersion))

			removed, err := addonManager.Remove(id, gameVersion)
			if err != nil {
				spinner.Fail(err.Error())
				return err
			}

			if removed {
				spinner.Succeed(fmt.Sprintf("Removed %s (%s) successfully", id, gameVersion))
			} else {
				spinner.Warn(fmt.Sprintf("%s (%s) not found", id, gameVersion))
			}

			return nil
		},
	}
	removeCmd.Flags().BoolP("retail", "r", true, "Remove from the retail version of the game")
	removeCmd.Flags().BoolP("classic", "c", false, "Remove from the classic version of the game")
	removeCmd.MarkFlagsMutuallyExclusive("classic", "retail")

	rootCmd.AddCommand(removeCmd)
}
