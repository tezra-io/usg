package claude

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	usageURL  = "https://api.anthropic.com/api/oauth/usage"
	betaHeader = "oauth-2025-04-20"
	userAgent  = "claude-code/2.1.0"
)

// UsageResponse matches the API response from api.anthropic.com/api/oauth/usage
type UsageResponse struct {
	FiveHour         *WindowData  `json:"five_hour"`
	SevenDay         *WindowData  `json:"seven_day"`
	SevenDayOAuthApps *WindowData `json:"seven_day_oauth_apps"`
	SevenDayOpus     *WindowData  `json:"seven_day_opus"`
	SevenDaySonnet   *WindowData  `json:"seven_day_sonnet"`
	IguanaNecktie    *WindowData  `json:"iguana_necktie"`
	ExtraUsage       *ExtraUsage  `json:"extra_usage"`
}

type WindowData struct {
	Utilization *float64 `json:"utilization"`
	ResetsAt    *string  `json:"resets_at"`
}

type ExtraUsage struct {
	IsEnabled    *bool    `json:"is_enabled"`
	MonthlyLimit *float64 `json:"monthly_limit"`
	UsedCredits  *float64 `json:"used_credits"`
	Utilization  *float64 `json:"utilization"`
	Currency     *string  `json:"currency"`
}

// Usage is the normalized output for display
type Usage struct {
	Session     *RateWindow `json:"session,omitempty"`
	Weekly      *RateWindow `json:"weekly,omitempty"`
	WeeklyOpus  *RateWindow `json:"weekly_opus,omitempty"`
	WeeklySonnet *RateWindow `json:"weekly_sonnet,omitempty"`
	ExtraUsage  *ExtraInfo  `json:"extra_usage,omitempty"`
}

type RateWindow struct {
	UsedPercent float64    `json:"used_percent"`
	ResetsAt    *time.Time `json:"resets_at,omitempty"`
}

type ExtraInfo struct {
	Enabled      bool    `json:"enabled"`
	MonthlyLimit float64 `json:"monthly_limit,omitempty"`
	UsedCredits  float64 `json:"used_credits,omitempty"`
	Utilization  float64 `json:"utilization,omitempty"`
	Currency     string  `json:"currency,omitempty"`
}

func FetchUsage() (*Usage, error) {
	token, err := GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}

	req, err := http.NewRequest("GET", usageURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-beta", betaHeader)
	req.Header.Set("User-Agent", userAgent)

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

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("unauthorized: token expired or invalid")
	}
	if resp.StatusCode == 403 {
		return nil, fmt.Errorf("forbidden: missing required scopes")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var raw UsageResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return normalize(&raw), nil
}

func normalize(raw *UsageResponse) *Usage {
	u := &Usage{}

	if raw.FiveHour != nil {
		u.Session = toRateWindow(raw.FiveHour)
	}
	if raw.SevenDay != nil {
		u.Weekly = toRateWindow(raw.SevenDay)
	}
	if raw.SevenDayOpus != nil {
		u.WeeklyOpus = toRateWindow(raw.SevenDayOpus)
	}
	if raw.SevenDaySonnet != nil {
		u.WeeklySonnet = toRateWindow(raw.SevenDaySonnet)
	}

	if raw.ExtraUsage != nil {
		ei := &ExtraInfo{}
		if raw.ExtraUsage.IsEnabled != nil {
			ei.Enabled = *raw.ExtraUsage.IsEnabled
		}
		if raw.ExtraUsage.MonthlyLimit != nil {
			ei.MonthlyLimit = *raw.ExtraUsage.MonthlyLimit
		}
		if raw.ExtraUsage.UsedCredits != nil {
			ei.UsedCredits = *raw.ExtraUsage.UsedCredits
		}
		if raw.ExtraUsage.Utilization != nil {
			ei.Utilization = *raw.ExtraUsage.Utilization
		}
		if raw.ExtraUsage.Currency != nil {
			ei.Currency = *raw.ExtraUsage.Currency
		}
		u.ExtraUsage = ei
	}

	return u
}

func toRateWindow(w *WindowData) *RateWindow {
	rw := &RateWindow{}
	if w.Utilization != nil {
		rw.UsedPercent = *w.Utilization // API returns percentage directly
	}
	if w.ResetsAt != nil {
		t, err := parseISO8601(*w.ResetsAt)
		if err == nil {
			rw.ResetsAt = &t
		}
	}
	return rw
}

func parseISO8601(s string) (time.Time, error) {
	// Try with fractional seconds first
	t, err := time.Parse("2006-01-02T15:04:05.999999999Z07:00", s)
	if err == nil {
		return t, nil
	}
	// Fallback without fractional seconds
	return time.Parse(time.RFC3339, s)
}
