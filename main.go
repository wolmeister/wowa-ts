package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// go build -ldflags "-X main.version=v1.0.0"
var (
	version = "dev"
	apiUrl  = "http://144.22.216.8:8888"
)

func getKeyValueStorePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	var path string
	switch runtime.GOOS {
	case "darwin":
		path = filepath.Join(homeDir, "Library", "Application Support", "wowa", "wowa.json")
	case "linux":
		path = filepath.Join(homeDir, ".config", "wowa", "wowa.json")
	case "windows":
		path = filepath.Join(homeDir, "AppData", "Roaming", "wowa", "wowa.json")
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return path, nil
}

func main() {
	var httpClient = NewHTTPClient()

	kvStorePath, err := getKeyValueStorePath()
	if err != nil {
		log.Fatal(err)
		return
	}

	kvStore, err := NewKeyValueStore(kvStorePath)
	if err != nil {
		log.Fatal(err)
		return
	}
	var configRepository = NewConfigRepository(kvStore)
	var userManager = NewUserManager(configRepository, apiUrl)
	var remoteAddonRepository = NewRemoteAddonRepository(userManager, apiUrl)
	var localAddonRepository = NewLocalAddonRepository(kvStore)

	curseToken, err := configRepository.Get(CurseToken)
	if err != nil {
		log.Fatal(err)
		return
	}

	var addonSearcher = NewAddonSearcher(httpClient, curseToken)
	var addonManager = NewAddonManager(addonSearcher, configRepository, localAddonRepository, remoteAddonRepository, httpClient)

	var rootCmd = &cobra.Command{
		Use:     "wowa",
		Short:   "World of Warcraft addon manager",
		Long:    `A simple CLI to manage World of Warcraft addons`,
		Version: version,
	}

	// Remove command
	var removeCmd = &cobra.Command{
		Use:     "rm <url>",
		Aliases: []string{"remove"},
		Short:   "Uninstall an addon",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Print: " + strings.Join(args, " "))
		},
	}
	removeCmd.Flags().BoolP("retail", "r", true, "Remove from the retail version of the game")
	removeCmd.Flags().BoolP("classic", "c", false, "Remove from the classic version of the game")
	removeCmd.MarkFlagsMutuallyExclusive("classic", "retail")

	// List command
	var lsCmd = &cobra.Command{
		Use:   "ls",
		Short: "List all installed addons",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Print: " + strings.Join(args, " "))
		},
	}

	// Backup command
	var backupCmd = &cobra.Command{
		Use:   "backup",
		Short: "Backup the WTF folder",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Print: " + strings.Join(args, " "))
		},
	}

	// Whoami command
	var whoamiCmd = &cobra.Command{
		Use:   "whoami",
		Short: "Display the user email currently logged in",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Print: " + strings.Join(args, " "))
		},
	}

	// Self-update command
	var selfUpdateCmd = &cobra.Command{
		Use:     "self-update",
		Aliases: []string{"su"},
		Short:   "Check and download new wowa updates",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Print: " + strings.Join(args, " "))
		},
	}

	// Export command
	var exportCmd = &cobra.Command{
		Use:   "export",
		Short: "Export all installed addons",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Print: " + strings.Join(args, " "))
		},
	}

	SetupAddCmd(rootCmd, addonManager)
	SetupUpdateCmd(rootCmd, addonManager, remoteAddonRepository)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(lsCmd)
	SetupConfigCmd(rootCmd, configRepository)
	rootCmd.AddCommand(backupCmd)
	SetupLoginCmd(rootCmd, userManager)
	rootCmd.AddCommand(whoamiCmd)
	rootCmd.AddCommand(selfUpdateCmd)
	rootCmd.AddCommand(exportCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
