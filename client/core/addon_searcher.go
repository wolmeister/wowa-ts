package core

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type AddonSearchResult struct {
	Slug        string
	Name        string
	Author      string
	GameVersion GameVersion
	Version     string
	Provider    AddonProvider
	ExternalId  string
	Url         string
	DownloadUrl RequestParams
}

type AddonSearcher struct {
	httpClient  *HTTPClient
	curseToken  string
	githubToken string
}

func NewAddonSearcher(httpClient *HTTPClient, curseToken string, githubToken string) *AddonSearcher {
	return &AddonSearcher{httpClient: httpClient, curseToken: curseToken, githubToken: githubToken}
}

func (as *AddonSearcher) parseCurseSlug(idOrUrl string) string {
	// Check if the ID starts with "cf:"
	if strings.HasPrefix(idOrUrl, "cf:") {
		return strings.TrimPrefix(idOrUrl, "cf:")
	}

	// Check if the ID is a CurseForge URL
	if strings.HasPrefix(idOrUrl, "https://www.curseforge.com/wow/addons/") {
		return strings.TrimPrefix(idOrUrl, "https://www.curseforge.com/wow/addons/")
	}

	// Check if the ID is a valid name (alphanumeric, -, or _)
	isValidName := regexp.MustCompile(`^[a-zA-Z0-9-_]+$`).MatchString
	if isValidName(idOrUrl) {
		return idOrUrl
	}

	return ""
}

func (as *AddonSearcher) curseSearch(slug string, gameVersion GameVersion) (AddonSearchResult, error) {

	type CurseModFileIndex struct {
		FileID            int `json:"fileId"`
		GameVersionTypeId int `json:"gameVersionTypeId"`
		ReleaseType       int `json:"releaseType"`
	}

	type CurseModAuthor struct {
		Name string `json:"name"`
	}

	type CurseMod struct {
		Id                 int                 `json:"id"`
		Slug               string              `json:"slug"`
		LatestFilesIndexes []CurseModFileIndex `json:"latestFilesIndexes"`
		Name               string              `json:"name"`
		Authors            []CurseModAuthor    `json:"authors"`
	}

	type SearchModsResponse struct {
		Data []CurseMod `json:"data"`
	}

	var gameVersionTypeId int
	switch gameVersion {
	case Retail:
		gameVersionTypeId = 517
	case Classic:
		gameVersionTypeId = 67408
	}

	curseHeaders := map[string]string{
		"x-api-key": as.curseToken,
	}

	var parsedSearchRes SearchModsResponse
	err := as.httpClient.Get(RequestParams{
		URL:     "https://api.curseforge.com/v1/mods/search",
		Headers: curseHeaders,
		Query: map[string]string{
			"gameId":            "1",
			"gameVersionTypeId": strconv.Itoa(gameVersionTypeId),
			"slug":              slug,
			"index":             "0",
			"sortField":         "2", // popularity
			"sortOrder":         "desc",
		},
	}, &parsedSearchRes)
	if err != nil {
		return AddonSearchResult{}, err
	}

	var curseMod *CurseMod
	for _, mod := range parsedSearchRes.Data {
		if mod.Slug == slug {
			curseMod = &mod
			break
		}
	}

	if curseMod == nil {
		return AddonSearchResult{}, errors.New("failed to find curse mod")
	}

	// Then, download the mod file
	var fileIndex *CurseModFileIndex
	for _, index := range curseMod.LatestFilesIndexes {
		if index.ReleaseType == 1 && index.GameVersionTypeId == gameVersionTypeId {
			fileIndex = &index
			break
		}
	}

	if fileIndex == nil {
		return AddonSearchResult{}, errors.New("failed to find curse mod file index")
	}

	type ModFile struct {
		DisplayName string `json:"displayName"`
		DownloadUrl string `json:"downloadUrl"`
	}

	type ModFileResponse struct {
		Data ModFile `json:"data"`
	}

	var parsedModFileRes ModFileResponse
	err = as.httpClient.Get(RequestParams{
		URL:     fmt.Sprintf("https://api.curseforge.com/v1/mods/%d/files/%d", curseMod.Id, fileIndex.FileID),
		Headers: curseHeaders,
	},
		&parsedModFileRes)
	if err != nil {
		return AddonSearchResult{}, err
	}

	modFile := parsedModFileRes.Data

	return AddonSearchResult{
		Slug:        slug,
		Name:        curseMod.Name,
		Author:      curseMod.Authors[0].Name,
		GameVersion: gameVersion,
		Version:     modFile.DisplayName,
		Provider:    Curse,
		ExternalId:  strconv.Itoa(curseMod.Id),
		Url:         fmt.Sprintf("https://www.curseforge.com/wow/addons/%s", curseMod.Slug),
		DownloadUrl: RequestParams{
			URL: modFile.DownloadUrl,
		},
	}, nil
}

func (as *AddonSearcher) parseGithubOrganizationAndRepository(idOrUrl string) (string, string) {
	rawRepo := ""

	// Check if the ID starts with "gh:"
	if strings.HasPrefix(idOrUrl, "gh:") {
		rawRepo = strings.TrimPrefix(idOrUrl, "cf:")
	}

	// Check if the ID is a CurseForge URL
	if strings.HasPrefix(idOrUrl, "https://github.com/") {
		rawRepo = strings.TrimPrefix(idOrUrl, "https://github.com/")
	}

	if rawRepo != "" {
		parts := strings.SplitN(rawRepo, "/", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}

	return "", ""
}

func (as *AddonSearcher) githubSearch(organization string, repository string, gameVersion GameVersion) (AddonSearchResult, error) {
	type GithubReleaseAsset struct {
		Id                 int    `json:"id"`
		Name               string `json:"name"`
		BrowserDownloadUrl string `json:"browser_download_url"`
	}

	type GithubRelease struct {
		TagName string               `json:"tag_name"`
		Assets  []GithubReleaseAsset `json:"assets"`
	}

	var latestRelease GithubRelease
	err := as.httpClient.Get(RequestParams{
		URL: fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", organization, repository),
		Headers: map[string]string{
			"Authorization": "token " + as.githubToken,
		},
	}, &latestRelease)

	if err != nil {
		return AddonSearchResult{}, err
	}

	var asset GithubReleaseAsset
	for _, a := range latestRelease.Assets {
		if strings.HasSuffix(a.BrowserDownloadUrl, ".zip") {
			asset = a
			break
		}
	}

	if asset.Id == 0 {
		return AddonSearchResult{}, errors.New("addon asset not found")

	}

	return AddonSearchResult{
		Slug:        repository,
		Name:        repository,
		Author:      organization,
		GameVersion: gameVersion,
		Version:     latestRelease.TagName,
		Provider:    Github,
		ExternalId:  fmt.Sprintf("%s/%s", organization, repository),
		Url:         fmt.Sprintf("https://github.com/%s/%s", organization, repository),
		DownloadUrl: RequestParams{
			URL: fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/assets/%d", organization, repository, asset.Id),
			Headers: map[string]string{
				"Accept":        "application/octet-stream",
				"Authorization": "token " + as.githubToken,
			},
		},
	}, nil
}

func (as *AddonSearcher) Search(idOrUrl string, gameVersion GameVersion) (AddonSearchResult, error) {
	curseSlug := as.parseCurseSlug(idOrUrl)
	if curseSlug != "" {
		return as.curseSearch(curseSlug, gameVersion)
	}

	organization, repository := as.parseGithubOrganizationAndRepository(idOrUrl)
	if organization != "" && repository != "" {
		return as.githubSearch(organization, repository, gameVersion)
	}

	return AddonSearchResult{}, errors.New("invalid addon id or url: " + idOrUrl)
}
