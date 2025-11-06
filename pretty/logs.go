package pretty

import (
	"fmt"
	"os"
	"strings"
)

// PrintLogs launches the TUI for viewing container logs
func PrintLogs(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: container name or ID required\n")
		fmt.Println("Usage: dockit logs [OPTIONS] CONTAINER")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  -f, --follow    Follow log output (stream new logs)")
		fmt.Println()
		fmt.Println("Interactive TUI Controls:")
		fmt.Println("  /               Start search")
		fmt.Println("  n / N           Jump to next/previous match")
		fmt.Println("  space           Pause/resume log streaming")
		fmt.Println("  ↑↓ / j k        Scroll up/down")
		fmt.Println("  PgUp / PgDn     Page up/down")
		fmt.Println("  g / G           Jump to top/bottom")
		fmt.Println("  q / Esc         Quit")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  dockit logs mycontainer          # View logs in interactive TUI")
		fmt.Println("  dockit logs -f mycontainer       # Follow logs with live updates")
		os.Exit(1)
	}

	// Parse arguments
	follow := false
	var containerID string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-f", "--follow":
			follow = true
		default:
			if !strings.HasPrefix(arg, "-") {
				containerID = arg
			}
		}
	}

	if containerID == "" {
		fmt.Fprintf(os.Stderr, "Error: container name or ID required\n")
		os.Exit(1)
	}

	// Launch TUI
	if err := LaunchLogsTUI(containerID, follow); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
