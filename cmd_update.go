package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wolmeister/wowa/spinny"
	"strings"
	"sync"
)

func SetupUpdateCmd(rootCmd *cobra.Command, addonManager *AddonManager, remoteAddonRepository *RemoteAddonRepository, weakAuraManager *WeakAuraManager) {
	var addCmd = &cobra.Command{
		Use:     "update",
		Short:   "Update all installed addons",
		Aliases: []string{"up"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: This should also remove uninstalled addons

			//TODO: Improve spinners/output logic. It's a mess right now. And ugly.
			var spinners = spinny.NewManager()
			spinners.Start()
			defer spinners.Stop()

			var addonsSpinner = spinners.NewSpinner("Updating addons...")
			var installedAddons []LocalAddon
			var updatedAddons []LocalAddon
			var reinstalledAddons []LocalAddon
			var errors []string

			addons, err := remoteAddonRepository.GetAddons()
			if err != nil {
				addonsSpinner.Fail("Failed to retrieve addons")
				return err
			}

			var wg sync.WaitGroup
			for _, addon := range addons {
				wg.Add(1)
				go func(addon RemoteAddon) {
					defer wg.Done()

					installResult, err := addonManager.Install(addon.Url, addon.GameVersion)
					if err != nil {
						errors = append(errors, fmt.Sprintf("Failed to update %s (%s) - %s", addon.Slug, addon.GameVersion, err.Error()))
						return
					}

					switch installResult.Status {
					case AddonInstallStatusAlreadyInstalled:
						// Do nothing
					case AddonInstallStatusInstalled:
						installedAddons = append(installedAddons, installResult.Addon)
					case AddonInstallStatusReinstalled:
						reinstalledAddons = append(reinstalledAddons, installResult.Addon)
					case AddonInstallStatusUpdated:
						updatedAddons = append(updatedAddons, installResult.Addon)
					}

				}(addon)
			}

			var waWg sync.WaitGroup
			waWg.Add(1)
			go func() {
				defer waWg.Done()
				waSpinner := spinners.NewSpinner("Updating weak auras...")
				waUpdates, err := weakAuraManager.UpdateAll(Retail)
				if err != nil {
					waSpinner.Fail("Failed to update retail weak auras")
				} else {
					if len(waUpdates) > 0 {
						waSpinner.Succeed(fmt.Sprintf("Updated %d retail weak auras", len(waUpdates)))
					} else {
						waSpinner.Info("All weak auras are up to date")
					}
				}
			}()

			wg.Wait()

			if len(errors) > 0 {
				addonsSpinner.Fail(fmt.Sprintf("Addons update complete with %d errors", len(errors)))
			} else {
				var messageParts []string
				if len(updatedAddons) > 0 {
					messageParts = append(messageParts, fmt.Sprintf("%d addons updated", len(updatedAddons)))
				}
				if len(reinstalledAddons) > 0 {
					messageParts = append(messageParts, fmt.Sprintf("%d addons reinstalled", len(reinstalledAddons)))
				}
				if len(installedAddons) > 0 {
					messageParts = append(messageParts, fmt.Sprintf("%d addons installed", len(installedAddons)))
				}

				if len(messageParts) > 0 {
					addonsSpinner.Succeed(strings.Join(messageParts, ", "))
				} else {
					addonsSpinner.Info("All addons are up to date")
				}
			}

			waWg.Wait()

			return nil
		},
	}

	rootCmd.AddCommand(addCmd)
}
