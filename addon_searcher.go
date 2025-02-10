package main

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
	DownloadUrl string
}

type AddonSearcher struct {
	httpClient *HTTPClient
	curseToken string
}

func NewAddonSearcher(httpClient *HTTPClient, curseToken string) *AddonSearcher {
	return &AddonSearcher{httpClient: httpClient, curseToken: curseToken}
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
		DownloadUrl: modFile.DownloadUrl,
	}, nil
}

func (as *AddonSearcher) Search(idOrUrl string, gameVersion GameVersion) (AddonSearchResult, error) {
	curseSlug := as.parseCurseSlug(idOrUrl)
	if curseSlug != "" {
		return as.curseSearch(curseSlug, gameVersion)
	}

	return AddonSearchResult{}, errors.New("invalid addon id or url: " + idOrUrl)
}
