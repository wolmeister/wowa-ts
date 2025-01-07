package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type RemoteAddon struct {
	Id          string        `json:"id"`
	UserId      string        `json:"user_id"`
	GameVersion GameVersion   `json:"game_version"`
	Slug        string        `json:"slug"`
	Name        string        `json:"name"`
	Author      string        `json:"author"`
	Provider    AddonProvider `json:"provider"`
	ExternalId  string        `json:"external_id"`
	Url         string        `json:"url"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type CreateAddonRequest struct {
	GameVersion GameVersion   `json:"game_version"`
	Slug        string        `json:"slug"`
	Name        string        `json:"name"`
	Author      string        `json:"author"`
	Provider    AddonProvider `json:"provider"`
	ExternalId  string        `json:"external_id"`
	Url         string        `json:"url"`
}

type RemoteAddonRepository struct {
	userManager *UserManager
	apiUrl      string
}

func NewRemoteAddonRepository(userManager *UserManager, apiUrl string) *RemoteAddonRepository {
	return &RemoteAddonRepository{userManager: userManager, apiUrl: apiUrl}
}

func (rar *RemoteAddonRepository) CreateAddon(addon CreateAddonRequest) (*RemoteAddon, error) {
	token, err := rar.userManager.GetUserToken()
	if err != nil || token == "" {
		return nil, errors.New("no user signed in")
	}

	body, err := json.Marshal(addon)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/addons", rar.apiUrl), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create addon: %s", resp.Status)
	}

	var remoteAddon RemoteAddon
	if err := json.NewDecoder(resp.Body).Decode(&remoteAddon); err != nil {
		return nil, err
	}

	return &remoteAddon, nil
}

func (rar *RemoteAddonRepository) DeleteAddon(slug string, gameVersion GameVersion) error {
	token, err := rar.userManager.GetUserToken()
	if err != nil || token == "" {
		return errors.New("no user signed in")
	}

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/addons/%s/%s", rar.apiUrl, gameVersion, slug), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete addon: %s", resp.Status)
	}

	return nil
}

func (rar *RemoteAddonRepository) GetAddons() ([]RemoteAddon, error) {
	token, err := rar.userManager.GetUserToken()
	if err != nil || token == "" {
		return nil, errors.New("no user signed in")
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/addons", rar.apiUrl), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get addons: %s", resp.Status)
	}

	var addons []RemoteAddon
	if err := json.NewDecoder(resp.Body).Decode(&addons); err != nil {
		return nil, err
	}

	return addons, nil
}

func (rar *RemoteAddonRepository) GetAddon(slug string, gameVersion GameVersion) (*RemoteAddon, error) {
	token, err := rar.userManager.GetUserToken()
	if err != nil || token == "" {
		return nil, errors.New("no user signed in")
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/addons/%s/%s", rar.apiUrl, gameVersion, slug), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Addon not found
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get addon: %s", resp.Status)
	}

	var remoteAddon RemoteAddon
	if err := json.NewDecoder(resp.Body).Decode(&remoteAddon); err != nil {
		return nil, err
	}

	return &remoteAddon, nil
}
