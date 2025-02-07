package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
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
	curseToken string
}

func NewAddonSearcher(curseToken string) *AddonSearcher {
	return &AddonSearcher{curseToken: curseToken}
}

func (ad *AddonSearcher) parseCurseSlug(idOrUrl string) string {
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

func (ad *AddonSearcher) curseSearch(slug string, gameVersion GameVersion) (AddonSearchResult, error) {
	var gameVersionTypeId int
	switch gameVersion {
	case Retail:
		gameVersionTypeId = 517
	case Classic:
		gameVersionTypeId = 67408
	}

	client := &http.Client{}

	// First, search for the mod
	searchURL, err := url.Parse("https://api.curseforge.com/v1/mods/search")
	if err != nil {
		return AddonSearchResult{}, err
	}

	query := searchURL.Query()
	query.Set("gameId", "1")
	query.Set("gameVersionTypeId", strconv.Itoa(gameVersionTypeId))
	query.Set("slug", slug)
	query.Set("index", "0")
	query.Set("sortField", "2") // popularity
	query.Set("sortOrder", "desc")
	searchURL.RawQuery = query.Encode()

	searchReq, err := http.NewRequest("GET", searchURL.String(), nil)
	if err != nil {
		return AddonSearchResult{}, err
	}
	searchReq.Header.Set("x-api-key", ad.curseToken)

	searchRes, err := client.Do(searchReq)
	if err != nil {
		return AddonSearchResult{}, err
	}
	defer func() {
		_ = searchRes.Body.Close()
	}()

	if searchRes.StatusCode != http.StatusOK {
		return AddonSearchResult{}, fmt.Errorf("failed to fetch data: %s", searchRes.Status)
	}

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

	var parsedSearchRes SearchModsResponse
	if err := json.NewDecoder(searchRes.Body).Decode(&parsedSearchRes); err != nil {
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

	modFileURL, err := url.Parse(fmt.Sprintf("https://api.curseforge.com/v1/mods/%d/files/%d", curseMod.Id, fileIndex.FileID))
	if err != nil {
		return AddonSearchResult{}, err
	}

	modFileReq, err := http.NewRequest("GET", modFileURL.String(), nil)
	if err != nil {
		return AddonSearchResult{}, err
	}
	modFileReq.Header.Set("x-api-key", ad.curseToken)

	modFileRes, err := client.Do(modFileReq)
	if err != nil {
		return AddonSearchResult{}, err
	}
	defer func() {
		_ = modFileRes.Body.Close()
	}()

	if modFileRes.StatusCode != http.StatusOK {
		return AddonSearchResult{}, fmt.Errorf("failed to fetch data: %s", searchRes.Status)
	}

	type ModFile struct {
		DisplayName string `json:"displayName"`
		DownloadUrl string `json:"downloadUrl"`
	}

	type ModFileResponse struct {
		Data ModFile `json:"data"`
	}

	var parsedModFileRes ModFileResponse
	if err := json.NewDecoder(modFileRes.Body).Decode(&parsedModFileRes); err != nil {
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

func (ad *AddonSearcher) Search(idOrUrl string, gameVersion GameVersion) (AddonSearchResult, error) {
	curseSlug := ad.parseCurseSlug(idOrUrl)
	if curseSlug != "" {
		return ad.curseSearch(curseSlug, gameVersion)
	}

	return AddonSearchResult{}, errors.New("invalid addon id or url: " + idOrUrl)
}
