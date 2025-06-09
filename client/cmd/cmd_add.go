package cmd

import (
	"fmt"
	"wowa/core"
	"wowa/spinny"

	"github.com/spf13/cobra"
)

func SetupAddCmd(rootCmd *cobra.Command, addonManager *core.AddonManager) {
	var addCmd = &cobra.Command{
		Use:   "add <url>",
		Short: "Install a new addon",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]

			var gameVersion core.GameVersion
			if cmd.Flag("retail").Value.String() == "true" {
				gameVersion = core.Retail
			} else {
				gameVersion = core.Classic
			}

			var spinners = spinny.NewManager()
			spinners.Start()
			defer spinners.Stop()

			var spinner = spinners.NewSpinner(fmt.Sprintf("Installing %s (%s)", url, gameVersion))

			installResult, err := addonManager.Install(url, gameVersion)
			if err != nil {
				spinner.Fail(err.Error())
				return err
			}

			switch installResult.Status {
			case core.AddonInstallStatusAlreadyInstalled:
				spinner.Info(fmt.Sprintf("%s (%s) %s is already installed", installResult.Addon.Slug, gameVersion, installResult.Addon.Version))
			case core.AddonInstallStatusInstalled:
				spinner.Succeed(fmt.Sprintf("%s (%s) %s installed successfully", installResult.Addon.Slug, gameVersion, installResult.Addon.Version))
			case core.AddonInstallStatusReinstalled:
				spinner.Warn(fmt.Sprintf("%s (%s) %s reinstalled", installResult.Addon.Slug, gameVersion, installResult.Addon.Version))
			case core.AddonInstallStatusUpdated:
				spinner.Info(fmt.Sprintf("%s (%s) updated to %s", installResult.Addon.Slug, gameVersion, installResult.Addon.Version))
			}

			return nil
		},
	}
	addCmd.Flags().BoolP("retail", "r", true, "Install in the retail version of the game")
	addCmd.Flags().BoolP("classic", "c", false, "Install in the classic version of the game")
	addCmd.MarkFlagsMutuallyExclusive("classic", "retail")

	rootCmd.AddCommand(addCmd)
}
