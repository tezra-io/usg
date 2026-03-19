package claude

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	refreshEndpoint = "https://console.anthropic.com/v1/oauth/token"
)

// CredentialsFile represents ~/.claude/.credentials.json or Keychain blob.
// Keychain uses camelCase "claudeAiOauth", credentials file uses snake_case "claude_ai_oauth".
type CredentialsFile struct {
	ClaudeAIOAuth      *OAuthEntry `json:"claude_ai_oauth,omitempty"`
	ClaudeAIOAuthCamel *OAuthEntry `json:"claudeAiOauth,omitempty"`
}

func (c *CredentialsFile) oauth() *OAuthEntry {
	if c.ClaudeAIOAuthCamel != nil {
		return c.ClaudeAIOAuthCamel
	}
	return c.ClaudeAIOAuth
}

type OAuthEntry struct {
	AccessToken   string   `json:"accessToken"`
	RefreshToken  string   `json:"refreshToken,omitempty"`
	ExpiresAt     float64  `json:"expiresAt,omitempty"` // milliseconds since epoch
	Scopes        []string `json:"scopes,omitempty"`
	RateLimitTier string   `json:"rateLimitTier,omitempty"`
}

type refreshResponse struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token,omitempty"`
	ExpiresIn    float64 `json:"expires_in,omitempty"` // seconds
}

// GetAccessToken tries Keychain first, then falls back to credentials file.
// If the token is expired, it attempts a refresh before returning.
func GetAccessToken() (string, error) {
	// Try macOS Keychain first
	token, err := readFromKeychain()
	if err == nil && token != "" {
		return token, nil
	}

	// Fallback to credentials file
	return readFromCredentialsFile()
}

// ForceRefresh forces a token refresh regardless of expiry.
// Used for 429 recovery when the token may have been revoked or expired
// between the expiry check and the API call.
func ForceRefresh() (string, error) {
	// Try keychain source first
	creds, err := readKeychainCreds()
	if err == nil {
		entry := creds.oauth()
		if entry != nil && entry.RefreshToken != "" {
			if err := refreshAndUpdate(entry, creds, "keychain"); err == nil {
				return entry.AccessToken, nil
			}
		}
	}

	// Try credentials file
	creds, credPath, err := readCredentialsFileCreds()
	if err == nil {
		entry := creds.oauth()
		if entry != nil && entry.RefreshToken != "" {
			if err := refreshAndUpdate(entry, creds, credPath); err == nil {
				return entry.AccessToken, nil
			}
		}
	}

	return "", fmt.Errorf("no refresh token available")
}

func isExpired(entry *OAuthEntry) bool {
	if entry.ExpiresAt <= 0 {
		return false // no expiry info, assume valid
	}
	expiresAtSec := entry.ExpiresAt / 1000 // ms to seconds
	return time.Now().Unix() >= int64(expiresAtSec)
}

func readKeychainCreds() (*CredentialsFile, error) {
	cmd := exec.Command("security", "find-generic-password",
		"-s", "Claude Code-credentials",
		"-w",
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("keychain: %w", err)
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, fmt.Errorf("keychain: empty value")
	}

	var creds CredentialsFile
	if err := json.Unmarshal([]byte(raw), &creds); err != nil {
		return nil, fmt.Errorf("keychain: %w", err)
	}

	return &creds, nil
}

func readFromKeychain() (string, error) {
	creds, err := readKeychainCreds()
	if err != nil {
		// Maybe it's just the token directly (non-JSON value)
		cmd := exec.Command("security", "find-generic-password",
			"-s", "Claude Code-credentials",
			"-w",
		)
		out, err2 := cmd.Output()
		if err2 != nil {
			return "", err
		}
		raw := strings.TrimSpace(string(out))
		if raw != "" {
			return raw, nil
		}
		return "", err
	}

	entry := creds.oauth()
	if entry == nil || entry.AccessToken == "" {
		return "", fmt.Errorf("keychain: no oauth entry found")
	}

	if isExpired(entry) && entry.RefreshToken != "" {
		_ = refreshAndUpdate(entry, creds, "keychain")
		// If refresh fails, fall through with existing token
	}

	return entry.AccessToken, nil
}

func readCredentialsFileCreds() (*CredentialsFile, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, "", fmt.Errorf("home dir: %w", err)
	}

	path := filepath.Join(home, ".claude", ".credentials.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("read %s: %w", path, err)
	}

	var creds CredentialsFile
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, "", fmt.Errorf("parse credentials: %w", err)
	}

	return &creds, path, nil
}

func readFromCredentialsFile() (string, error) {
	creds, path, err := readCredentialsFileCreds()
	if err != nil {
		return "", err
	}

	entry := creds.oauth()
	if entry == nil {
		return "", fmt.Errorf("no oauth entry in credentials")
	}
	if entry.AccessToken == "" {
		return "", fmt.Errorf("empty access token in credentials")
	}

	if isExpired(entry) && entry.RefreshToken != "" {
		_ = refreshAndUpdate(entry, creds, path)
	}

	return entry.AccessToken, nil
}

func refreshAndUpdate(entry *OAuthEntry, creds *CredentialsFile, source string) error {
	refreshed, err := doRefresh(entry.RefreshToken)
	if err != nil {
		return err
	}

	entry.AccessToken = refreshed.AccessToken
	if refreshed.RefreshToken != "" {
		entry.RefreshToken = refreshed.RefreshToken
	}
	if refreshed.ExpiresIn > 0 {
		entry.ExpiresAt = float64(time.Now().UnixMilli()) + refreshed.ExpiresIn*1000
	}

	if source == "keychain" {
		writeToKeychain(creds)
	} else {
		writeToCredentialsFile(creds, source)
	}

	return nil
}

func doRefresh(refreshToken string) (*refreshResponse, error) {
	payload := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
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

func writeToKeychain(creds *CredentialsFile) {
	data, err := json.Marshal(creds)
	if err != nil {
		return
	}

	cmd := exec.Command("security", "add-generic-password",
		"-U",
		"-s", "Claude Code-credentials",
		"-a", "Claude Code",
		"-w", string(data),
	)
	_ = cmd.Run()
}

func writeToCredentialsFile(creds *CredentialsFile, path string) {
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0600)
}
