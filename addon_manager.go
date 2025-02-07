package main

import "fmt"

type AddonManager struct {
	//userManager *UserManager
	addonSearcher *AddonSearcher
}

type AddonInstallResult struct{}

func NewAddonManager(addonSearcher *AddonSearcher) *AddonManager {
	return &AddonManager{addonSearcher: addonSearcher}
}

func (am *AddonManager) Install(url string, gameVersion GameVersion) (AddonInstallResult, error) {
	searchResult, err := am.addonSearcher.Search(url, gameVersion)
	if err != nil {
		return AddonInstallResult{}, err
	}
	fmt.Println(searchResult)
	return AddonInstallResult{}, nil
}
