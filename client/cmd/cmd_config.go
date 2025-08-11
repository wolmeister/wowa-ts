package cmd

import (
	"errors"
	"fmt"
	"wowa/core"
	"wowa/utils"

	"github.com/spf13/cobra"
)

func SetupConfigCmd(rootCmd *cobra.Command, configRepository *core.ConfigRepository) {
	var configCmd = &cobra.Command{
		Use:   "config <key> [value]",
		Short: "Manage configuration",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			switch core.Config(key) {
			case core.CurseToken, core.GithubToken, core.GameDir, core.AuthToken:
				break
			default:
				// TODO: Add the available keys to the error message.
				return errors.New("the config key is not allowed")
			}

			if len(args) > 1 {
				err := configRepository.Set(core.Config(key), &args[1])
				if err != nil {
					return err
				}
			} else {
				value, err := configRepository.Get(core.Config(key))
				if err != nil {
					return err
				}

				if value == "" {
					fmt.Println(utils.AnsiBlue + "<null>" + utils.AnsiReset)
				} else {
					fmt.Println(value)
				}
			}

			return nil
		},
	}

	rootCmd.AddCommand(configCmd)
}
