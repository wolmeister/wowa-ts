package main

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
)

func AddConfigCmd(rootCmd *cobra.Command, configRepository *ConfigRepository) {
	var configCmd = &cobra.Command{
		Use:   "config <key> [value]",
		Short: "Manage configuration",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			switch Config(key) {
			case CurseToken, GameDir, AuthToken:
				break
			default:
				// TODO: Add the available keys to the error message.
				return errors.New("the config key is not allowed")
			}

			if len(args) > 1 {
				err := configRepository.Set(Config(key), &args[1])
				if err != nil {
					return err
				}
			} else {
				value, err := configRepository.Get(Config(key))
				if err != nil {
					return err
				}

				if value == "" {
					fmt.Println(AnsiBlue + "<null>" + AnsiReset)
				} else {
					fmt.Println(value)
				}
			}

			return nil
		},
	}

	rootCmd.AddCommand(configCmd)
}
