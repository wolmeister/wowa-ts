package main

import (
	"errors"
	"fmt"
	"path/filepath"
)

type AddonManager struct {
	//userManager *UserManager
	configRepository     *ConfigRepository
	addonSearcher        *AddonSearcher
	localAddonRepository *LocalAddonRepository
}

type AddonInstallResult struct{}

func NewAddonManager(addonSearcher *AddonSearcher, localAddonRepository *LocalAddonRepository) *AddonManager {
	return &AddonManager{addonSearcher: addonSearcher, localAddonRepository: localAddonRepository}
}

func (am *AddonManager) getAddonsFolder(gameVersion GameVersion) (string, error) {
	gameDir, err := am.configRepository.Get(GameDir)
	if err != nil {
		return "", err
	}
	if gameDir == "" {
		return "", errors.New("game dir is not defined")
	}

	var versionFolder string
	if gameVersion == "classic" {
		versionFolder = "_classic_era_"
	} else {
		versionFolder = "_retail_"
	}

	addonsFolder := filepath.Join(gameDir, versionFolder, "Interface", "AddOns")

	return addonsFolder, nil
}

func (am *AddonManager) isAddonInstallationValid(localAddon *LocalAddon) (bool, error) {
	// TODO
	return false, nil
}

func (am *AddonManager) Install(url string, gameVersion GameVersion) (AddonInstallResult, error) {
	searchResult, err := am.addonSearcher.Search(url, gameVersion)
	if err != nil {
		return AddonInstallResult{}, err
	}

	existingAddon, err := am.localAddonRepository.Get(searchResult.Slug, gameVersion)
	if err != nil {
		return AddonInstallResult{}, err
	}

	if existingAddon != nil && existingAddon.Version == searchResult.Version {
		isInstallationValid, err := am.isAddonInstallationValid(existingAddon)
		if err != nil {
			return AddonInstallResult{}, err
		}
		if !isInstallationValid {
			return AddonInstallResult{}, nil
		}
	}

	fmt.Println(searchResult)

	return AddonInstallResult{}, nil
}
