package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/tezra-io/usg/internal/claude"
	"github.com/tezra-io/usg/internal/codex"
	"github.com/tezra-io/usg/internal/output"
)

const version = "0.1.0"

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printUsageHelp()
		os.Exit(1)
	}

	var (
		jsonMode  bool
		watchMode bool
		watchSec  = 30
		command   string
	)

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json", "-j":
			jsonMode = true
		case "--watch", "-w":
			watchMode = true
			if i+1 < len(args) {
				if n, err := strconv.Atoi(args[i+1]); err == nil && n > 0 {
					watchSec = n
					i++
				}
			}
		case "--version", "-v":
			fmt.Println("usg " + version)
			return
		case "--help", "-h":
			printUsageHelp()
			return
		default:
			if command == "" && args[i][0] != '-' {
				command = args[i]
			}
		}
	}

	if command == "" {
		command = "all"
	}

	run := func() {
		switch command {
		case "claude":
			runClaude(jsonMode)
		case "codex":
			runCodex(jsonMode)
		case "all":
			runAll(jsonMode)
		default:
			fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
			printUsageHelp()
			os.Exit(1)
		}
	}

	if watchMode {
		for {
			run()
			time.Sleep(time.Duration(watchSec) * time.Second)
			if !jsonMode {
				fmt.Print("\033[2J\033[H") // clear screen
			}
		}
	} else {
		run()
	}
}

func runClaude(jsonMode bool) {
	u, err := claude.FetchUsage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "claude: %v\n", err)
		if jsonMode {
			fmt.Println(`{"claude": null}`)
		}
		return
	}
	if jsonMode {
		output.PrintClaudeJSON(u)
	} else {
		output.PrintClaude(u)
	}
}

func runCodex(jsonMode bool) {
	u, err := codex.FetchUsage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "codex: %v\n", err)
		if jsonMode {
			fmt.Println(`{"codex": null}`)
		}
		return
	}
	if jsonMode {
		output.PrintCodexJSON(u)
	} else {
		output.PrintCodex(u)
	}
}

func runAll(jsonMode bool) {
	cu, cerr := claude.FetchUsage()
	xu, xerr := codex.FetchUsage()

	if jsonMode {
		output.PrintAllJSON(cu, xu)
		return
	}

	if cerr != nil {
		fmt.Fprintf(os.Stderr, "claude: %v\n", cerr)
	} else {
		output.PrintClaude(cu)
	}

	if xerr != nil {
		fmt.Fprintf(os.Stderr, "codex: %v\n", xerr)
	} else {
		if cerr == nil {
			fmt.Println()
		}
		output.PrintCodex(xu)
	}
}

func printUsageHelp() {
	fmt.Fprintln(os.Stderr, `usg - Claude Code & Codex usage reporter

Usage:
  usg <command> [flags]

Commands:
  claude    Show Claude Code usage
  codex     Show Codex usage
  all       Show both (default)

Flags:
  --json, -j          JSON output
  --watch, -w [secs]  Poll mode (default: 30s)
  --version, -v       Show version
  --help, -h          Show help`)
}
