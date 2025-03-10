package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type SelfUpdateResult struct {
	Updated     bool
	FromVersion string
	ToVersion   string
}

type SelfUpdateManager struct {
	currentVersion string
	httpClient     *HTTPClient
}

func NewSelfUpdateManager(currentVersion string, httpClient *HTTPClient) *SelfUpdateManager {
	return &SelfUpdateManager{
		currentVersion: currentVersion,
		httpClient:     httpClient,
	}
}

func (sum *SelfUpdateManager) getLatestGithubRelease(owner, repo string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to get latest release")
	}

	var release map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return nil, err
	}

	return release, nil
}

func (sum *SelfUpdateManager) isVersionAtOrBelow(version1 string, version2 string) bool {
	// TODO: Improve this
	v1 := strings.Split(version1[1:], ".")
	v2 := strings.Split(version2[1:], ".")

	for i := 0; i < len(v1); i++ {
		if i >= len(v2) {
			return false
		}

		p1, _ := strconv.Atoi(v1[i])
		p2, _ := strconv.Atoi(v2[i])

		if p1 != p2 {
			return p1 < p2
		}
	}

	return true
}

func (sum *SelfUpdateManager) UpdateToLatest() (SelfUpdateResult, error) {
	type GithubReleaseAsset struct {
		Name               string `json:"name"`
		BrowserDownloadUrl string `json:"browser_download_url"`
	}

	type GithubRelease struct {
		TagName string               `json:"tag_name"`
		Assets  []GithubReleaseAsset `json:"assets"`
	}

	var latestRelease GithubRelease
	err := sum.httpClient.Get(RequestParams{
		URL: "https://api.github.com/repos/wolmeister/wowa-ts/releases/latest",
	}, &latestRelease)
	if err != nil {
		return SelfUpdateResult{}, err
	}

	latestVersion := latestRelease.TagName
	if sum.isVersionAtOrBelow(latestVersion, version) {
		return SelfUpdateResult{Updated: false}, nil
	}

	platform := runtime.GOOS
	var asset GithubReleaseAsset
	for _, a := range latestRelease.Assets {
		switch platform {
		case "windows":
			if a.Name == "wowa-win64.exe" {
				asset = a
				break
			}
		case "linux":
			if a.Name == "wowa-linux64" {
				asset = a
				break
			}
		default:
			return SelfUpdateResult{}, errors.New("unsupported operating system")
		}
	}

	downloadedRelease, err := sum.httpClient.GetBytes(RequestParams{URL: asset.BrowserDownloadUrl})
	if err != nil {
		return SelfUpdateResult{}, err
	}

	executablePath, err := os.Executable()
	if err != nil {
		return SelfUpdateResult{}, err
	}

	backupPath := executablePath + ".backup"
	err = os.Rename(executablePath, backupPath)
	if err != nil {
		return SelfUpdateResult{}, err
	}

	err = os.WriteFile(executablePath, downloadedRelease, 0777)
	if err != nil {
		return SelfUpdateResult{}, err
	}

	return SelfUpdateResult{Updated: true, FromVersion: sum.currentVersion, ToVersion: latestVersion}, nil
}
