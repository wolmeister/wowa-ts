package core

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type UserManager struct {
	configRepository *ConfigRepository
	apiUrl           string
}

func NewUserManager(configRepository *ConfigRepository, apiUrl string) *UserManager {
	return &UserManager{
		configRepository: configRepository,
		apiUrl:           apiUrl,
	}
}

func (um *UserManager) GetUserToken() (string, error) {
	return um.configRepository.Get("auth.token")
}

func (um *UserManager) GetUserEmail() (string, error) {
	token, err := um.GetUserToken()
	if err != nil {
		return "", err
	}
	if token == "" {
		return "", nil
	}

	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid token format")
	}

	base64Payload := parts[1]
	base64MissingPadding := (4 - len(base64Payload)%4) % 4
	if base64MissingPadding > 0 {
		base64Payload += strings.Repeat("=", base64MissingPadding)
	}

	payloadBytes, err := base64.StdEncoding.DecodeString(base64Payload)
	if err != nil {
		return "", err
	}

	var payload map[string]interface{}
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		return "", err
	}

	email, ok := payload["email"].(string)
	if !ok {
		return "", fmt.Errorf("email not found in token payload")
	}
	return email, nil
}

func (um *UserManager) SignIn(email, password string) error {
	body, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/login", um.apiUrl), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusBadRequest {
			return fmt.Errorf("invalid email or password")
		}
		return fmt.Errorf("failed to sign in: %s", resp.Status)
	}

	rawToken, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	token := string(rawToken)

	return um.configRepository.Set("auth.token", &token)
}
