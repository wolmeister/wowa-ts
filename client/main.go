package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"wowa/cmd"
	"wowa/core"
	"wowa/gui"

	"github.com/spf13/cobra"
)

// go build -ldflags "-X main.version=FOO"
var (
	version = "dev"
	apiUrl  = ""
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
	var httpClient = core.NewHTTPClient()

	kvStorePath, err := getKeyValueStorePath()
	if err != nil {
		log.Fatal(err)
		return
	}

	kvStore, err := core.NewKeyValueStore(kvStorePath)
	if err != nil {
		log.Fatal(err)
		return
	}

	var configRepository = core.NewConfigRepository(kvStore)
	var userManager = core.NewUserManager(configRepository, apiUrl)
	var remoteAddonRepository = core.NewRemoteAddonRepository(userManager, apiUrl)
	var localAddonRepository = core.NewLocalAddonRepository(kvStore)

	curseToken, err := configRepository.Get(core.CurseToken)
	if err != nil {
		log.Fatal(err)
		return
	}

	var addonSearcher = core.NewAddonSearcher(httpClient, curseToken)
	var addonManager = core.NewAddonManager(addonSearcher, configRepository, localAddonRepository, remoteAddonRepository, httpClient)
	var selfUpdateManager = core.NewSelfUpdateManager(version, httpClient)
	var weakAuraManager = core.NewWeakAuraManager(configRepository, httpClient)

	if len(os.Args) == 1 {
		err := gui.Start(localAddonRepository)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	var rootCmd = &cobra.Command{
		Use:     "wowa",
		Short:   "World of Warcraft addon manager",
		Long:    `A simple CLI to manage World of Warcraft addons`,
		Version: version,
	}

	cmd.SetupAddCmd(rootCmd, addonManager)
	cmd.SetupUpdateCmd(rootCmd, addonManager, remoteAddonRepository, weakAuraManager)
	cmd.SetupRemoveCmd(rootCmd, addonManager)
	cmd.SetupLsCmd(rootCmd, localAddonRepository)
	cmd.SetupConfigCmd(rootCmd, configRepository)
	cmd.SetupLoginCmd(rootCmd, userManager)
	cmd.SetupWhoamiCmd(rootCmd, userManager)
	cmd.SetupSelfUpdateCmd(rootCmd, selfUpdateManager)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
