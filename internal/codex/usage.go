package codex

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://chatgpt.com/backend-api"
	usagePath      = "/wham/usage"
)

// APIResponse matches the chatgpt.com/backend-api/wham/usage response
type APIResponse struct {
	PlanType  *string    `json:"plan_type"`
	RateLimit *RateLimit `json:"rate_limit"`
	Credits   *Credits   `json:"credits"`
}

type RateLimit struct {
	PrimaryWindow   *Window `json:"primary_window"`
	SecondaryWindow *Window `json:"secondary_window"`
}

type Window struct {
	UsedPercent        int `json:"used_percent"`
	ResetAt            int `json:"reset_at"` // unix timestamp
	LimitWindowSeconds int `json:"limit_window_seconds"`
}

type Credits struct {
	HasCredits bool        `json:"has_credits"`
	Unlimited  bool        `json:"unlimited"`
	Balance    json.Number `json:"balance"`
}

// Usage is the normalized output for display
type Usage struct {
	PlanType  string      `json:"plan_type,omitempty"`
	Primary   *RateWindow `json:"primary,omitempty"`
	Secondary *RateWindow `json:"secondary,omitempty"`
	Credits   *CreditInfo `json:"credits,omitempty"`
}

type RateWindow struct {
	UsedPercent    int        `json:"used_percent"`
	ResetsAt       *time.Time `json:"resets_at,omitempty"`
	WindowMinutes  int        `json:"window_minutes,omitempty"`
}

type CreditInfo struct {
	HasCredits bool    `json:"has_credits"`
	Unlimited  bool    `json:"unlimited"`
	Balance    float64 `json:"balance,omitempty"`
}

func FetchUsage() (*Usage, error) {
	token, accountID, err := GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}

	baseURL := resolveBaseURL()
	url := baseURL + usagePath

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "CodexBar")
	if accountID != "" {
		req.Header.Set("ChatGPT-Account-Id", accountID)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, fmt.Errorf("unauthorized: token expired or invalid (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var raw APIResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return normalize(&raw), nil
}

func normalize(raw *APIResponse) *Usage {
	u := &Usage{}

	if raw.PlanType != nil {
		u.PlanType = *raw.PlanType
	}

	if raw.RateLimit != nil {
		if raw.RateLimit.PrimaryWindow != nil {
			w := raw.RateLimit.PrimaryWindow
			rw := &RateWindow{
				UsedPercent:   w.UsedPercent,
				WindowMinutes: w.LimitWindowSeconds / 60,
			}
			if w.ResetAt > 0 {
				t := time.Unix(int64(w.ResetAt), 0)
				rw.ResetsAt = &t
			}
			u.Primary = rw
		}
		if raw.RateLimit.SecondaryWindow != nil {
			w := raw.RateLimit.SecondaryWindow
			rw := &RateWindow{
				UsedPercent:   w.UsedPercent,
				WindowMinutes: w.LimitWindowSeconds / 60,
			}
			if w.ResetAt > 0 {
				t := time.Unix(int64(w.ResetAt), 0)
				rw.ResetsAt = &t
			}
			u.Secondary = rw
		}
	}

	if raw.Credits != nil {
		ci := &CreditInfo{
			HasCredits: raw.Credits.HasCredits,
			Unlimited:  raw.Credits.Unlimited,
		}
		if bal, err := raw.Credits.Balance.Float64(); err == nil {
			ci.Balance = bal / 100 // cents to dollars
		}
		u.Credits = ci
	}

	return u
}

func resolveBaseURL() string {
	codexHome := os.Getenv("CODEX_HOME")
	if codexHome == "" {
		home, _ := os.UserHomeDir()
		if home != "" {
			codexHome = filepath.Join(home, ".codex")
		}
	}

	if codexHome != "" {
		configPath := filepath.Join(codexHome, "config.toml")
		data, err := os.ReadFile(configPath)
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "chatgpt_base_url") {
					parts := strings.SplitN(line, "=", 2)
					if len(parts) == 2 {
						url := strings.TrimSpace(parts[1])
						url = strings.Trim(url, "\"'")
						return normalizeBaseURL(url)
					}
				}
			}
		}
	}

	return defaultBaseURL
}

func normalizeBaseURL(url string) string {
	url = strings.TrimRight(url, "/")
	if url == "https://chatgpt.com" || url == "https://chat.openai.com" {
		return url + "/backend-api"
	}
	if !strings.HasSuffix(url, "/backend-api") {
		return url + "/backend-api"
	}
	return url
}
