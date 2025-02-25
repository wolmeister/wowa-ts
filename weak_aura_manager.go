package main

import (
	"errors"
	"fmt"
	lua "github.com/yuin/gopher-lua"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type WeakAuraManager struct {
	configRepository *ConfigRepository
	httpClient       *HTTPClient
}

type LocalWeakAura struct {
	Name    string
	Slug    string
	Version int
}

type WeakAuraUpdate struct {
	Slug        string
	Name        string
	Author      string
	WagoVersion int
	WagoSemver  string
	Encoded     string
}

func NewWeakAuraManager(configRepository *ConfigRepository, httpClient *HTTPClient) *WeakAuraManager {
	return &WeakAuraManager{
		configRepository: configRepository,
		httpClient:       httpClient,
	}
}

// TODO: Move this to another file. This is duplicated in the addon manager file.
func (wam *WeakAuraManager) getGameVersionFolder(gameVersion GameVersion) (string, error) {
	gameDir, err := wam.configRepository.Get(GameDir)
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

	addonsFolder := filepath.Join(gameDir, versionFolder)

	return addonsFolder, nil
}

func (wam *WeakAuraManager) getWeakAurasLuaPath(gameVersion GameVersion) ([]string, error) {
	var luaPaths []string

	gameVersionFolder, err := wam.getGameVersionFolder(gameVersion)
	if err != nil {
		return luaPaths, err
	}

	accountsFolder := filepath.Join(gameVersionFolder, "WTF", "Account")
	_, err = os.Stat(accountsFolder)
	if err != nil {
		if os.IsNotExist(err) {
			return luaPaths, nil
		}
		return luaPaths, err
	}

	accountsFolders, err := os.ReadDir(accountsFolder)
	if err != nil {
		return luaPaths, err
	}

	for _, accountFolder := range accountsFolders {
		if !accountFolder.IsDir() || accountFolder.Name() == "SavedVariables" {
			continue
		}
		luaPath := filepath.Join(accountsFolder, accountFolder.Name(), "SavedVariables", "WeakAuras.lua")
		_, err = os.Stat(luaPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return luaPaths, err
		}
		luaPaths = append(luaPaths, luaPath)
	}

	return luaPaths, nil
}

func (wam *WeakAuraManager) parseWeakAuraFile(luaPath string) ([]LocalWeakAura, error) {
	wagoRegex := regexp.MustCompile(`^https://wago\.io/([a-zA-Z0-9]+)/(\d+)$`)
	var weakAuras []LocalWeakAura

	L := lua.NewState()
	defer L.Close()

	if err := L.DoFile(luaPath); err != nil {
		return nil, err
	}

	weakAurasTable := L.GetGlobal("WeakAurasSaved")
	if weakAurasTable.Type() != lua.LTTable {
		return nil, errors.New("invalid WeakAuras structure")
	}

	rawDisplaysTable := L.GetField(weakAurasTable, "displays")
	if rawDisplaysTable.Type() != lua.LTTable {
		return nil, errors.New("invalid WeakAuras structure")
	}
	displaysTable, ok := rawDisplaysTable.(*lua.LTable)
	if !ok {
		return nil, errors.New("displays is not a valid Lua table")
	}

	L.ForEach(displaysTable, func(key lua.LValue, value lua.LValue) {
		if value.Type() != lua.LTTable {
			return
		}

		weakAuraDisplay, ok := value.(*lua.LTable)
		if !ok {
			return
		}

		// Skip children of each weak aura group
		if weakAuraDisplay.RawGetString("parent").Type() != lua.LTNil {
			return
		}

		maybeUrl := weakAuraDisplay.RawGetString("url")
		if maybeUrl.Type() != lua.LTString {
			return
		}

		urlMatch := wagoRegex.FindStringSubmatch(maybeUrl.String())
		if urlMatch == nil {
			return
		}

		version, err := strconv.Atoi(urlMatch[2])
		if err != nil {
			return
		}

		weakAuras = append(weakAuras, LocalWeakAura{
			Name:    key.String(),
			Slug:    urlMatch[1],
			Version: version,
		})
	})

	return weakAuras, nil
}

func (wam *WeakAuraManager) getInstalledWeakAuras(gameVersion GameVersion) ([]LocalWeakAura, error) {
	var weakAuras []LocalWeakAura
	weakAurasNames := NewSet[string]()

	luaPaths, err := wam.getWeakAurasLuaPath(gameVersion)
	if err != nil {
		return weakAuras, err
	}

	for _, luaPath := range luaPaths {
		accountWeakAuras, err := wam.parseWeakAuraFile(luaPath)
		if err != nil {
			return weakAuras, err
		}

		for _, accountWeakAura := range accountWeakAuras {
			if !weakAurasNames.Contains(accountWeakAura.Name) {
				weakAurasNames.Add(accountWeakAura.Name)
				weakAuras = append(weakAuras, accountWeakAura)
			}
		}
	}

	return weakAuras, nil
}

func (wam *WeakAuraManager) generateCompanionDataFile(weakAuraUpdates []WeakAuraUpdate) string {
	lines := []string{
		"WowaCompanionData = {",
		"   WeakAuras = {",
		"       slugs = {",
	}

	for _, update := range weakAuraUpdates {
		// TODO: Support changelogs
		changelog := ""
		line := fmt.Sprintf(
			"[\"%s\"] = {\n"+
				"    name = [=[%s]=],\n"+
				"    author = [=[%s]=],\n"+
				"    encoded = [=[%s]=],\n"+
				"    wagoVersion = [=[%d]=],\n"+
				"    wagoSemver = [=[%s]=],\n"+
				"    source = [=[%s]=],\n"+
				"    versionNote = [=[%s]=],\n"+
				"},",
			update.Slug, update.Name, update.Author, update.Encoded, update.WagoVersion, update.WagoSemver, "Wago", changelog,
		)

		lines = append(lines, line)
	}

	lines = append(lines, "       }")
	lines = append(lines, "   }")
	lines = append(lines, "}")

	return strings.Join(lines, "\n")
}

func (wam *WeakAuraManager) installCompanionAddon(updates []WeakAuraUpdate, gameVersion GameVersion) error {
	// Compute the addon path
	gameVersionFolder, err := wam.getGameVersionFolder(gameVersion)
	if err != nil {
		return err
	}
	addonFolder := filepath.Join(gameVersionFolder, "Interface", "AddOns", "WowaCompanion")

	// Create the addon directory
	if err := os.MkdirAll(addonFolder, os.ModePerm); err != nil {
		return err
	}

	// Create the Data.lua file
	dataLuaPath := filepath.Join(addonFolder, "Data.lua")
	err = os.WriteFile(dataLuaPath, []byte(wam.generateCompanionDataFile(updates)), os.ModePerm)
	if err != nil {
		return err
	}

	// Create WowaCompanion.toc file
	wowaCompanionTocPath := filepath.Join(addonFolder, "WowaCompanion.toc")
	tocContent := []string{
		"## Title: Wowa Companion",
		"## Author: Victor Wolmeister",
		"## Version: 1.0.0",
		"## Notes: Wowa Companion addon to keep things up to date",
		"## DefaultState: Enabled",
		"## OptionalDeps: WeakAuras",
		"",
		"Data.lua",
		"WowaCompanion.lua",
	}
	err = os.WriteFile(wowaCompanionTocPath, []byte(strings.Join(tocContent, "\n")), os.ModePerm)
	if err != nil {
		return err
	}

	// Create WowaCompanion.lua file
	wowaCompanionLuaPath := filepath.Join(addonFolder, "WowaCompanion.lua")
	luaContent := []string{
		"local frame = CreateFrame(\"FRAME\")",
		"frame:RegisterEvent(\"ADDON_LOADED\")",
		"frame:SetScript(\"OnEvent\", function(_, _, addonName)",
		"   if addonName == \"WowaCompanion\" then",
		"       if WeakAuras and WeakAuras.AddCompanionData and WowaCompanionData then",
		"           local WeakAurasData = WowaCompanionData.WeakAuras",
		"           if WeakAurasData then",
		"                WeakAuras.AddCompanionData(WeakAurasData)",
		"           end",
		"       end",
		"   end",
		"end)",
	}

	err = os.WriteFile(wowaCompanionLuaPath, []byte(strings.Join(luaContent, "\n")), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func (wam *WeakAuraManager) UpdateAll(gameVersion GameVersion) ([]WeakAuraUpdate, error) {
	weakAuras, err := wam.getInstalledWeakAuras(gameVersion)
	if err != nil {
		return nil, err
	}

	type WagoCheckUpdatesRequest struct {
		Ids []string `json:"ids"`
	}
	type WagoCheckUpdatesRequestResponse struct {
		Slug        string `json:"slug"`
		Name        string `json:"name"`
		Author      string `json:"username"`
		WagoVersion int    `json:"version"`
		WagoSemver  string `json:"versionString"`
	}

	var wagoResponse []WagoCheckUpdatesRequestResponse
	var wagoRequest WagoCheckUpdatesRequest

	for _, wa := range weakAuras {
		wagoRequest.Ids = append(wagoRequest.Ids, wa.Slug)
	}

	err = wam.httpClient.Post(RequestParams{
		URL: "https://data.wago.io/api/check/weakauras",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, wagoRequest, &wagoResponse)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var updates []WeakAuraUpdate

	for _, waUpdate := range wagoResponse {
		shouldUpdate := false
		for _, wa := range weakAuras {
			if wa.Slug == waUpdate.Slug {
				shouldUpdate = waUpdate.WagoVersion > wa.Version
				break
			}
		}
		if !shouldUpdate {
			continue
		}

		wg.Add(1)
		go func(waUpdate WagoCheckUpdatesRequestResponse) {
			defer wg.Done()

			encodedBytes, err := wam.httpClient.GetBytes(RequestParams{
				URL: "https://data.wago.io/api/raw/encoded?id=" + waUpdate.Slug,
			})
			if err != nil {
				// TODO: Handle errors
				wg.Done()
				return
			}
			updates = append(updates, WeakAuraUpdate{
				Slug:        waUpdate.Slug,
				Name:        waUpdate.Name,
				Author:      waUpdate.Author,
				WagoVersion: waUpdate.WagoVersion,
				WagoSemver:  waUpdate.WagoSemver,
				Encoded:     string(encodedBytes),
			})

		}(waUpdate)
	}

	wg.Wait()

	err = wam.installCompanionAddon(updates, gameVersion)
	if err != nil {
		return nil, err
	}

	return updates, nil
}
