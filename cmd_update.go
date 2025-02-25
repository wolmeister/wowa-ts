package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wolmeister/wowa/spinny"
	"sync"
)

func SetupUpdateCmd(rootCmd *cobra.Command, addonManager *AddonManager, remoteAddonRepository *RemoteAddonRepository, weakAuraManager *WeakAuraManager) {
	var addCmd = &cobra.Command{
		Use:     "update",
		Short:   "Update all installed addons",
		Aliases: []string{"up"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: This should also remove uninstalled addons

			addons, err := remoteAddonRepository.GetAddons()
			if err != nil {
				return err
			}

			var spinners = spinny.NewManager()
			spinners.Start()
			defer spinners.Stop()

			var wg sync.WaitGroup
			for _, addon := range addons {
				wg.Add(1)
				go func(addon RemoteAddon) {
					defer wg.Done()

					var spinner = spinners.NewSpinner(fmt.Sprintf("Updating %s (%s)", addon.Slug, addon.GameVersion))

					installResult, err := addonManager.Install(addon.Url, addon.GameVersion)
					if err != nil {
						spinner.Fail(fmt.Sprintf("Failed to update %s (%s) - %s", addon.Slug, addon.GameVersion, err.Error()))
						return
					}

					switch installResult.Status {
					case AddonInstallStatusAlreadyInstalled:
						spinner.Info(fmt.Sprintf("%s (%s) %s is already up to date", installResult.Addon.Slug, installResult.Addon.GameVersion, installResult.Addon.Version))
					case AddonInstallStatusInstalled:
						spinner.Succeed(fmt.Sprintf("%s (%s) %s installed successfully", installResult.Addon.Slug, installResult.Addon.GameVersion, installResult.Addon.Version))
					case AddonInstallStatusReinstalled:
						spinner.Warn(fmt.Sprintf("%s (%s) %s reinstalled", installResult.Addon.Slug, installResult.Addon.GameVersion, installResult.Addon.Version))
					case AddonInstallStatusUpdated:
						spinner.Succeed(fmt.Sprintf("%s (%s) updated to %s", installResult.Addon.Slug, installResult.Addon.GameVersion, installResult.Addon.Version))
					}

				}(addon)
			}

			waSpinner := spinners.NewSpinner("Updating retail weak auras")
			waUpdates, err := weakAuraManager.UpdateAll(Retail)
			if err != nil {
				waSpinner.Fail("Failed to update retail weak auras")
			} else {
				if len(waUpdates) > 0 {
					waSpinner.Succeed(fmt.Sprintf("Updated %d retail weak auras", len(waUpdates)))
				} else {
					waSpinner.Info("No weak aura to update")
				}
			}

			wg.Wait()

			return nil
		},
	}

	rootCmd.AddCommand(addCmd)
}
