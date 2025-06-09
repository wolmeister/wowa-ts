package cmd

import (
	"fmt"
	"strings"
	"wowa/core"

	"github.com/spf13/cobra"
)

func SetupLsCmd(rootCmd *cobra.Command, localAddonRepository *core.LocalAddonRepository) {
	var addCmd = &cobra.Command{
		Use:   "ls",
		Short: "List all installed addons",
		RunE: func(cmd *cobra.Command, args []string) error {
			addons, err := localAddonRepository.GetAll(nil)
			if err != nil {
				return err
			}

			table := [][]string{{"ID", "Name", "Version", "Game Version", "Updated at"}}
			largestColumns := []int{len("ID"), len("Name"), len("Version"), len("Game Version"), len("Updated at")}

			for _, addon := range addons {
				addonName := strings.Split(addon.Name, " ... ")[0]
				updatedAt := addon.UpdatedAt.Format("2006-01-02 15:04:05")

				table = append(table, []string{addon.Id, addonName, addon.Version, string(addon.GameVersion), updatedAt})
				largestColumns[0] = max(len(addon.Id), largestColumns[0])
				largestColumns[1] = max(len(addonName), largestColumns[1])
				largestColumns[2] = max(len(addon.Version), largestColumns[2])
				largestColumns[3] = max(len(addon.GameVersion), largestColumns[3])
				largestColumns[4] = max(len(updatedAt), largestColumns[4])
			}

			// Print top border
			for _, largestColumn := range largestColumns {
				for i := 0; i < largestColumn+2; i++ {
					fmt.Printf("-")
				}
			}
			fmt.Printf("-\n")

			// Print table
			for _, row := range table {
				for columnIndex, column := range row {
					fmt.Printf("| %s", column)
					for i := len(column); i < largestColumns[columnIndex]; i++ {
						fmt.Printf(" ")
					}
					if columnIndex == len(row)-1 {
						fmt.Printf("|")
					}
				}
				fmt.Printf("\n")
			}

			// Print bottom border
			for _, largestColumn := range largestColumns {
				for i := 0; i < largestColumn+2; i++ {
					fmt.Printf("-")
				}
			}
			fmt.Printf("-\n")

			return nil
		},
	}

	rootCmd.AddCommand(addCmd)
}
