package main

import (
	"encoding/json"
	"time"
)

type GameVersion string

const (
	Retail  GameVersion = "retail"
	Classic GameVersion = "classic"
)

type AddonProvider string

const (
	Curse AddonProvider = "curse"
)

type LocalAddon struct {
	Id          string        `json:"id"`
	Name        string        `json:"name"`
	Slug        string        `json:"slug"`
	Author      string        `json:"author"`
	Version     string        `json:"version"`
	GameVersion GameVersion   `json:"gameVersion"`
	Directories []string      `json:"directories"`
	Provider    AddonProvider `json:"provider"`
	ExternalId  string        `json:"providerId"`
	UpdatedAt   time.Time     `json:"updatedAt"`
}

type LocalAddonRepository struct {
	kvStore *KeyValueStore
}

func NewLocalAddonRepository(kvStore *KeyValueStore) *LocalAddonRepository {
	return &LocalAddonRepository{kvStore: kvStore}
}

func (lar *LocalAddonRepository) Save(addon LocalAddon) error {
	data, err := json.Marshal(addon)
	if err != nil {
		return err
	}
	stringData := string(data)
	return lar.kvStore.Set([]string{"local-addons", string(addon.GameVersion), addon.Id}, &stringData)
}

func (lar *LocalAddonRepository) Delete(id string, gameVersion GameVersion) error {
	return lar.kvStore.Set([]string{"local-addons", string(gameVersion), id}, nil)
}

func (lar *LocalAddonRepository) Get(id string, gameVersion GameVersion) (*LocalAddon, error) {
	data, err := lar.kvStore.Get([]string{"local-addons", string(gameVersion), id})
	if err != nil {
		return nil, err
	}
	if data == "" {
		// Not found
		return nil, nil
	}
	var addon LocalAddon
	if err := json.Unmarshal([]byte(data), &addon); err != nil {
		return nil, err
	}
	return &addon, nil
}

func (lar *LocalAddonRepository) GetAll(gameVersion *GameVersion) ([]LocalAddon, error) {
	var key []string
	if gameVersion != nil {
		key = []string{"local-addons", string(*gameVersion)}
	} else {
		key = []string{"local-addons"}
	}

	dataList, err := lar.kvStore.GetByPrefix(key)
	if err != nil {
		return nil, err
	}

	var addons []LocalAddon
	for _, jsonData := range dataList {
		var addon LocalAddon
		if err := json.Unmarshal([]byte(jsonData), &addon); err != nil {
			return nil, err
		}
		addons = append(addons, addon)
	}

	return addons, nil
}
