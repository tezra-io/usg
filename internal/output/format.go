package output

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tezra-io/usg/internal/claude"
	"github.com/tezra-io/usg/internal/codex"
)

// CombinedOutput is the JSON-mode output structure
type CombinedOutput struct {
	Claude *claude.Usage `json:"claude,omitempty"`
	Codex  *codex.Usage  `json:"codex,omitempty"`
}

func PrintClaudeJSON(u *claude.Usage) {
	out := CombinedOutput{Claude: u}
	printJSON(out)
}

func PrintCodexJSON(u *codex.Usage) {
	out := CombinedOutput{Codex: u}
	printJSON(out)
}

func PrintAllJSON(cu *claude.Usage, xu *codex.Usage) {
	out := CombinedOutput{Claude: cu, Codex: xu}
	printJSON(out)
}

func printJSON(v any) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}

func PrintClaude(u *claude.Usage) {
	fmt.Println("Claude Code")
	fmt.Println(strings.Repeat("─", 40))

	if u.Session != nil {
		fmt.Printf("  Session (5h):   %5.1f%% used", u.Session.UsedPercent)
		if u.Session.ResetsAt != nil {
			fmt.Printf("  resets %s", formatReset(*u.Session.ResetsAt))
		}
		fmt.Println()
	}

	if u.Weekly != nil {
		fmt.Printf("  Weekly (7d):    %5.1f%% used", u.Weekly.UsedPercent)
		if u.Weekly.ResetsAt != nil {
			fmt.Printf("  resets %s", formatReset(*u.Weekly.ResetsAt))
		}
		fmt.Println()
	}

	if u.WeeklyOpus != nil {
		fmt.Printf("  Weekly Opus:    %5.1f%% used", u.WeeklyOpus.UsedPercent)
		if u.WeeklyOpus.ResetsAt != nil {
			fmt.Printf("  resets %s", formatReset(*u.WeeklyOpus.ResetsAt))
		}
		fmt.Println()
	}

	if u.WeeklySonnet != nil {
		fmt.Printf("  Weekly Sonnet:  %5.1f%% used", u.WeeklySonnet.UsedPercent)
		if u.WeeklySonnet.ResetsAt != nil {
			fmt.Printf("  resets %s", formatReset(*u.WeeklySonnet.ResetsAt))
		}
		fmt.Println()
	}

	if u.ExtraUsage != nil && u.ExtraUsage.Enabled {
		currency := u.ExtraUsage.Currency
		if currency == "" {
			currency = "USD"
		}
		fmt.Printf("  Extra usage:    $%.2f / $%.2f %s (%.1f%%)\n",
			u.ExtraUsage.UsedCredits, u.ExtraUsage.MonthlyLimit,
			currency, u.ExtraUsage.Utilization)
	}
}

func PrintCodex(u *codex.Usage) {
	fmt.Println("Codex")
	fmt.Println(strings.Repeat("─", 40))

	if u.PlanType != "" {
		fmt.Printf("  Plan: %s\n", u.PlanType)
	}

	if u.Primary != nil {
		label := "Primary"
		if u.Primary.WindowMinutes > 0 {
			label = fmt.Sprintf("Primary (%s)", formatDuration(u.Primary.WindowMinutes))
		}
		fmt.Printf("  %-16s %3d%% used", label, u.Primary.UsedPercent)
		if u.Primary.ResetsAt != nil {
			fmt.Printf("  resets %s", formatReset(*u.Primary.ResetsAt))
		}
		fmt.Println()
	}

	if u.Secondary != nil {
		label := "Secondary"
		if u.Secondary.WindowMinutes > 0 {
			label = fmt.Sprintf("Secondary (%s)", formatDuration(u.Secondary.WindowMinutes))
		}
		fmt.Printf("  %-16s %3d%% used", label, u.Secondary.UsedPercent)
		if u.Secondary.ResetsAt != nil {
			fmt.Printf("  resets %s", formatReset(*u.Secondary.ResetsAt))
		}
		fmt.Println()
	}

	if u.Credits != nil {
		if u.Credits.Unlimited {
			fmt.Println("  Credits: unlimited")
		} else if u.Credits.HasCredits {
			fmt.Printf("  Credits: $%.2f\n", u.Credits.Balance)
		} else {
			fmt.Println("  Credits: none")
		}
	}
}

func formatReset(t time.Time) string {
	d := time.Until(t)
	if d < 0 {
		return "now"
	}
	if d < time.Minute {
		return fmt.Sprintf("in %ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("in %dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		if m > 0 {
			return fmt.Sprintf("in %dh%dm", h, m)
		}
		return fmt.Sprintf("in %dh", h)
	}
	return fmt.Sprintf("in %dd%dh", int(d.Hours())/24, int(d.Hours())%24)
}

func formatDuration(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	h := minutes / 60
	m := minutes % 60
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh%dm", h, m)
}
