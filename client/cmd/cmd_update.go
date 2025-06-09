package cmd

import (
	"fmt"
	"os"
	"sync"
	"wowa/core"
	"wowa/utils"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

func SetupUpdateCmd(rootCmd *cobra.Command, addonManager *core.AddonManager, remoteAddonRepository *core.RemoteAddonRepository, weakAuraManager *core.WeakAuraManager) {
	var addCmd = &cobra.Command{
		Use:     "update",
		Short:   "Update all installed addons",
		Aliases: []string{"up"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: This should also remove uninstalled addons

			var messages []string
			var wg sync.WaitGroup

			progressBar := progressbar.NewOptions(
				-1,
				progressbar.OptionSetDescription("Updating addons and weak auras..."),
				progressbar.OptionShowCount(),
				progressbar.OptionOnCompletion(func() {
					fmt.Fprint(os.Stderr, "\n\n")

					for _, message := range messages {
						fmt.Fprintf(os.Stderr, " -> %s\n", message)
					}
				}),
				progressbar.OptionSetTheme(progressbar.Theme{
					Saucer:        "#",
					SaucerHead:    ">",
					SaucerPadding: " ",
					BarStart:      "[",
					BarEnd:        "]",
				}))
			defer progressBar.Finish()

			addons, err := remoteAddonRepository.GetAddons()
			if err != nil {
				progressBar.Describe("Failed to retrieve addons")
				return err
			}

			// Update addons
			progressBar.ChangeMax(len(addons))
			wg.Add(len(addons))

			for _, addon := range addons {
				go func(addon core.RemoteAddon) {
					defer wg.Done()
					defer progressBar.Add(1)

					installResult, err := addonManager.Install(addon.Url, addon.GameVersion)
					if err != nil {
						messages = append(messages, fmt.Sprintf("%sFailed to update addon %s (%s) - %s %s", utils.AnsiRed, addon.Slug, addon.GameVersion, err.Error(), utils.AnsiReset))
						return
					}

					switch installResult.Status {
					case core.AddonInstallStatusAlreadyInstalled:
						// Do nothing
					case core.AddonInstallStatusInstalled:
						messages = append(messages, fmt.Sprintf("Addon %s (%s) %s installed", addon.Slug, addon.GameVersion, installResult.Addon.Version))
					case core.AddonInstallStatusReinstalled:
						messages = append(messages, fmt.Sprintf("Addon %s (%s) %s reinstalled", addon.Slug, addon.GameVersion, installResult.Addon.Version))
					case core.AddonInstallStatusUpdated:
						messages = append(messages, fmt.Sprintf("Addon %s (%s) updated to %s", addon.Slug, addon.GameVersion, installResult.Addon.Version))
					}

				}(addon)
			}

			// Update weak auras
			progressBar.AddMax(1)
			wg.Add(1)

			go func() {
				defer wg.Done()
				defer progressBar.Add(1)

				waUpdates, err := weakAuraManager.UpdateAll(core.Retail)
				if err != nil {
					messages = append(messages, fmt.Sprintf("%sFailed to update weak auras %s %s", utils.AnsiRed, err.Error(), utils.AnsiReset))
					return
				}

				for _, waUpdate := range waUpdates {
					messages = append(messages, fmt.Sprintf("Weak Aura %s updated to %s", waUpdate.Name, waUpdate.WagoSemver))
				}
			}()

			wg.Wait()

			if len(messages) == 0 {
				messages = append(messages, "All addons and weak auras are up to date!")
			}

			return nil
		},
	}

	rootCmd.AddCommand(addCmd)
}
