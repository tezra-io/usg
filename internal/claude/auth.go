package claude

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken,omitempty"`
	ExpiresAt    float64  `json:"expiresAt,omitempty"` // milliseconds since epoch
	Scopes       []string `json:"scopes,omitempty"`
	RateLimitTier string  `json:"rateLimitTier,omitempty"`
}

// GetAccessToken tries Keychain first, then falls back to credentials file.
func GetAccessToken() (string, error) {
	// Try macOS Keychain first
	token, err := readFromKeychain()
	if err == nil && token != "" {
		return token, nil
	}

	// Fallback to credentials file
	return readFromCredentialsFile()
}

func readFromKeychain() (string, error) {
	cmd := exec.Command("security", "find-generic-password",
		"-s", "Claude Code-credentials",
		"-w",
	)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("keychain: %w", err)
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return "", fmt.Errorf("keychain: empty value")
	}

	// The keychain value is the JSON credentials blob
	var creds CredentialsFile
	if err := json.Unmarshal([]byte(raw), &creds); err != nil {
		// Maybe it's just the token directly
		return raw, nil
	}

	if entry := creds.oauth(); entry != nil && entry.AccessToken != "" {
		return entry.AccessToken, nil
	}

	return "", fmt.Errorf("keychain: no oauth entry found")
}

func readFromCredentialsFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}

	path := filepath.Join(home, ".claude", ".credentials.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}

	var creds CredentialsFile
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", fmt.Errorf("parse credentials: %w", err)
	}

	entry := creds.oauth()
	if entry == nil {
		return "", fmt.Errorf("no oauth entry in credentials")
	}

	if entry.AccessToken == "" {
		return "", fmt.Errorf("empty access token in credentials")
	}

	return entry.AccessToken, nil
}
