package codex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	refreshEndpoint = "https://auth.openai.com/oauth/token"
	clientID        = "app_EMoamEEZ73f0CkXaXp7hrann"
	refreshScope    = "openai profile email"
	refreshMaxAge   = 8 * 24 * time.Hour // 8 days
)

// AuthFile represents ~/.codex/auth.json
type AuthFile struct {
	Tokens      *Tokens `json:"tokens"`
	LastRefresh string  `json:"last_refresh,omitempty"`
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token,omitempty"`
	AccountID    string `json:"account_id,omitempty"`
}

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token,omitempty"`
}

// GetAccessToken reads the Codex OAuth token, refreshing if stale.
func GetAccessToken() (token string, accountID string, err error) {
	codexHome := os.Getenv("CODEX_HOME")
	if codexHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", "", fmt.Errorf("home dir: %w", err)
		}
		codexHome = filepath.Join(home, ".codex")
	}

	authPath := filepath.Join(codexHome, "auth.json")
	data, err := os.ReadFile(authPath)
	if err != nil {
		return "", "", fmt.Errorf("read %s: %w", authPath, err)
	}

	var auth AuthFile
	if err := json.Unmarshal(data, &auth); err != nil {
		return "", "", fmt.Errorf("parse auth.json: %w", err)
	}

	if auth.Tokens == nil {
		return "", "", fmt.Errorf("no tokens in auth.json")
	}

	if auth.Tokens.AccessToken == "" {
		return "", "", fmt.Errorf("empty access_token in auth.json")
	}

	// Check if refresh is needed (> 8 days since last refresh)
	needsRefresh := false
	if auth.LastRefresh != "" {
		lastRefresh, err := time.Parse(time.RFC3339, auth.LastRefresh)
		if err == nil && time.Since(lastRefresh) > refreshMaxAge {
			needsRefresh = true
		}
	}

	if needsRefresh && auth.Tokens.RefreshToken != "" {
		refreshed, err := doRefresh(auth.Tokens.RefreshToken)
		if err == nil {
			auth.Tokens.AccessToken = refreshed.AccessToken
			if refreshed.RefreshToken != "" {
				auth.Tokens.RefreshToken = refreshed.RefreshToken
			}
			if refreshed.IDToken != "" {
				auth.Tokens.IDToken = refreshed.IDToken
			}
			auth.LastRefresh = time.Now().UTC().Format(time.RFC3339)

			// Write back
			updated, _ := json.MarshalIndent(auth, "", "  ")
			if updated != nil {
				_ = os.WriteFile(authPath, updated, 0600)
			}
		}
		// If refresh fails, try with existing token anyway
	}

	return auth.Tokens.AccessToken, auth.Tokens.AccountID, nil
}

func doRefresh(refreshToken string) (*refreshResponse, error) {
	payload := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     clientID,
		"scope":         refreshScope,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", refreshEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("refresh failed %d: %s", resp.StatusCode, string(respBody))
	}

	var result refreshResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse refresh response: %w", err)
	}

	return &result, nil
}
