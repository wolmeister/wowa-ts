package core

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"wowa/utils"
)

type AddonManager struct {
	addonSearcher         *AddonSearcher
	configRepository      *ConfigRepository
	localAddonRepository  *LocalAddonRepository
	remoteAddonRepository *RemoteAddonRepository
	httpClient            *HTTPClient
}

type AddonInstallStatus int

const (
	AddonInstallStatusAlreadyInstalled AddonInstallStatus = iota
	AddonInstallStatusInstalled
	AddonInstallStatusReinstalled
	AddonInstallStatusUpdated
)

type AddonInstallResult struct {
	Addon  LocalAddon
	Status AddonInstallStatus
}

func NewAddonManager(addonSearcher *AddonSearcher, configRepository *ConfigRepository, localAddonRepository *LocalAddonRepository, remoteAddonRepository *RemoteAddonRepository, httpClient *HTTPClient) *AddonManager {
	return &AddonManager{addonSearcher: addonSearcher, configRepository: configRepository, localAddonRepository: localAddonRepository, remoteAddonRepository: remoteAddonRepository, httpClient: httpClient}
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
	// Get the folder where the addons are installed
	addonsFolder, err := am.getAddonsFolder(localAddon.GameVersion)
	if err != nil {
		return false, err
	}

	var wg sync.WaitGroup
	valid := make(chan bool, len(localAddon.Directories))
	for _, addonDirectory := range localAddon.Directories {
		wg.Add(1)
		go func(addonDirectory string) {
			defer wg.Done()
			modulePath := filepath.Join(addonsFolder, addonDirectory)
			_, err := os.Stat(modulePath)
			if err != nil {
				valid <- false
			} else {
				valid <- true
			}
		}(addonDirectory)
	}

	go func() {
		wg.Wait()
		close(valid)
	}()

	for v := range valid {
		if !v {
			return false, nil
		}
	}

	return true, nil
}

func (am *AddonManager) extractAddon(zipBytes []byte, addonsFolder string) ([]string, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return nil, err
	}

	rootDirectories := utils.NewSet[string]()
	allDirectories := utils.NewSet[string]()

	// First, discover all directories
	for _, file := range zipReader.File {
		if file.Mode().IsDir() {
			continue
		}

		// Check for ZipSlip
		if strings.Contains(file.Name, "..") || filepath.IsAbs(file.Name) {
			return nil, fmt.Errorf("illegal file path: %s", file.Name)
		}

		pathParts := strings.Split(file.Name, "/")
		rootDirectories.Add(pathParts[0])
		allDirectories.Add(filepath.Dir(file.Name))
	}

	// Then, create all directories
	for _, dir := range allDirectories.ToArray() {
		err := os.MkdirAll(filepath.Join(addonsFolder, dir), 0755)
		if err != nil {
			return nil, err
		}
	}

	// And finally, extract all the files
	extractAndWriteFile := func(file *zip.File) error {
		if file.Mode().IsDir() {
			return nil
		}

		// Open the file
		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer func(fileReader io.ReadCloser) {
			_ = fileReader.Close()
		}(fileReader)

		// Extract the file
		path := filepath.Join(addonsFolder, file.Name)
		outputFile, err := os.Create(path)
		if err != nil {
			return err
		}
		defer func(outputFile *os.File) {
			_ = outputFile.Close()
		}(outputFile)

		_, err = io.Copy(outputFile, fileReader)
		if err != nil {
			return err
		}

		return nil
	}

	for _, file := range zipReader.File {
		err := extractAndWriteFile(file)
		if err != nil {
			return nil, err
		}
	}

	return rootDirectories.ToArray(), nil
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

	// Get the folder where the addons are installed
	addonsFolder, err := am.getAddonsFolder(gameVersion)
	if err != nil {
		return AddonInstallResult{}, err
	}

	// Check if the addon is already installed
	if existingAddon != nil && existingAddon.Version == searchResult.Version {
		// Check if the addon installation is valid. We should reinstall if something is missing,
		isInstallationValid, err := am.isAddonInstallationValid(existingAddon)
		if err != nil {
			return AddonInstallResult{}, err
		}
		if isInstallationValid {
			return AddonInstallResult{
				Addon:  *existingAddon,
				Status: AddonInstallStatusAlreadyInstalled,
			}, nil
		}
	}

	// Check if the addons folder does NOT already exist
	if _, err := os.Stat(addonsFolder); os.IsNotExist(err) {
		// Create the folder
		err := os.MkdirAll(addonsFolder, os.ModePerm)
		if err != nil {
			return AddonInstallResult{}, err
		}
	} else if existingAddon != nil {
		// If the folder already exists, and the addon is already installed locally
		// we need to delete all the related files.
		for _, d := range existingAddon.Directories {
			dirPath := filepath.Join(addonsFolder, d)
			err := os.RemoveAll(dirPath)
			if err != nil {
				return AddonInstallResult{}, err
			}
		}
	}

	// Download the zip
	zipBytes, err := am.httpClient.GetBytes(RequestParams{URL: searchResult.DownloadUrl})
	if err != nil {
		return AddonInstallResult{}, err
	}

	// Extract the zip
	rootDirectories, err := am.extractAddon(zipBytes, addonsFolder)
	if err != nil {
		return AddonInstallResult{}, err
	}

	// Save the addon to the remote repository
	remoteAddon, err := am.remoteAddonRepository.GetAddon(searchResult.Slug, gameVersion)
	if err != nil {
		return AddonInstallResult{}, err
	}
	if remoteAddon == nil {
		_, err := am.remoteAddonRepository.CreateAddon(CreateAddonRequest{
			Slug:        searchResult.Slug,
			GameVersion: gameVersion,
			Author:      searchResult.Author,
			Name:        searchResult.Name,
			Provider:    searchResult.Provider,
			ExternalId:  searchResult.ExternalId,
			Url:         searchResult.Url,
		})
		if err != nil {
			return AddonInstallResult{}, err
		}
	}

	// Save the addon to the local repository
	installedAddon := LocalAddon{
		Id:          searchResult.Slug,
		Slug:        searchResult.Slug,
		GameVersion: gameVersion,
		Name:        searchResult.Name,
		Version:     searchResult.Version,
		Author:      searchResult.Author,
		Directories: rootDirectories,
		Provider:    searchResult.Provider,
		ExternalId:  searchResult.ExternalId,
		UpdatedAt:   time.Now(),
	}
	err = am.localAddonRepository.Save(installedAddon)
	if err != nil {
		return AddonInstallResult{}, err
	}

	resultStatus := AddonInstallStatusInstalled
	if existingAddon != nil && existingAddon.Version == searchResult.Version {
		resultStatus = AddonInstallStatusReinstalled
	} else if existingAddon != nil {
		resultStatus = AddonInstallStatusUpdated
	}

	return AddonInstallResult{
		Addon:  installedAddon,
		Status: resultStatus,
	}, nil
}

func (am *AddonManager) Remove(id string, gameVersion GameVersion) (bool, error) {
	// TODO: This should check if the local addons are up to date with the remote repository.

	//Get the local addon
	localAddon, err := am.localAddonRepository.Get(id, gameVersion)
	if err != nil {
		return false, err
	}
	if localAddon == nil {
		return false, nil
	}

	// Get the folder where the addons are installed
	addonsFolder, err := am.getAddonsFolder(gameVersion)
	if err != nil {
		return false, err
	}

	// Delete the addon files
	for _, d := range localAddon.Directories {
		dirPath := filepath.Join(addonsFolder, d)
		err := os.RemoveAll(dirPath)
		if err != nil {
			return false, err
		}
	}

	//Delete the local addon
	err = am.localAddonRepository.Delete(id, gameVersion)
	if err != nil {
		return false, err
	}

	// Delete the remote addon
	err = am.remoteAddonRepository.DeleteAddon(localAddon.Slug, gameVersion)
	if err != nil {
		return false, err
	}

	return true, nil
}
