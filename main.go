package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"runtime"
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
	var selfUpdateManager = NewSelfUpdateManager(version, httpClient)

	var rootCmd = &cobra.Command{
		Use:     "wowa",
		Short:   "World of Warcraft addon manager",
		Long:    `A simple CLI to manage World of Warcraft addons`,
		Version: version,
	}

	// Backup command
	var backupCmd = &cobra.Command{
		Use:   "backup",
		Short: "Backup the WTF folder",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("TODO: This will be implemented again soon")
		},
	}

	// Export command
	var exportCmd = &cobra.Command{
		Use:   "export",
		Short: "Export all installed addons",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("TODO: This will be implemented again soon")
		},
	}

	SetupAddCmd(rootCmd, addonManager)
	SetupUpdateCmd(rootCmd, addonManager, remoteAddonRepository)
	SetupRemoveCmd(rootCmd, addonManager)
	SetupLsCmd(rootCmd, localAddonRepository)
	SetupConfigCmd(rootCmd, configRepository)
	rootCmd.AddCommand(backupCmd)
	SetupLoginCmd(rootCmd, userManager)
	SetupWhoamiCmd(rootCmd, userManager)
	SetupSelfUpdateCmd(rootCmd, selfUpdateManager)
	rootCmd.AddCommand(exportCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
