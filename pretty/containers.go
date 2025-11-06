package pretty

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/fatih/color"
)

var (
	green  = color.New(color.FgGreen, color.Bold)
	red    = color.New(color.FgRed, color.Bold)
	yellow = color.New(color.FgYellow, color.Bold)
	cyan   = color.New(color.FgCyan, color.Bold)
	blue   = color.New(color.FgBlue, color.Bold)
	gray   = color.New(color.FgHiBlack)
)

// PrintContainers displays containers in a pretty format
func PrintContainers(args []string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Docker client: %v\n", err)
		os.Exit(1)
	}
	defer cli.Close()

	ctx := context.Background()

	// Check if -a flag is present for showing all containers
	showAll := false
	for _, arg := range args {
		if arg == "-a" || arg == "--all" {
			showAll = true
			break
		}
	}

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: showAll})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing containers: %v\n", err)
		os.Exit(1)
	}

	if len(containers) == 0 {
		gray.Println("No containers found")
		if !showAll {
			gray.Println("(use 'dockit ps -a' to see all containers)")
		}
		return
	}

	// Print header
	fmt.Println()
	cyan.Println("CONTAINERS")
	cyan.Println(strings.Repeat("─", 90))

	// Print containers
	for _, c := range containers {
		// Status indicator and color
		var statusColor *color.Color
		var indicator string
		if c.State == "running" {
			statusColor = green
			indicator = "●"
		} else if c.State == "exited" {
			statusColor = gray
			indicator = "○"
		} else if c.State == "paused" {
			statusColor = yellow
			indicator = "⏸"
		} else {
			statusColor = red
			indicator = "✖"
		}

		// Container ID (short)
		containerID := c.ID
		if len(containerID) > 12 {
			containerID = containerID[:12]
		}
		idWidth := 12
		idPadded := containerID + strings.Repeat(" ", idWidth-len(containerID))

		// Container name
		name := strings.TrimPrefix(c.Names[0], "/")
		nameWidth := 30
		if len(name) > nameWidth {
			name = name[:nameWidth-3] + "..."
		}
		namePadded := name + strings.Repeat(" ", nameWidth-len(name))

		// Image name
		image := c.Image
		imageWidth := 30
		if len(image) > imageWidth {
			image = image[:imageWidth-3] + "..."
		}
		imagePadded := image + strings.Repeat(" ", imageWidth-len(image))

		// State
		stateWidth := 10
		statePadded := c.State + strings.Repeat(" ", stateWidth-len(c.State))

		// Print main line
		statusColor.Print(indicator)
		fmt.Print(" ")
		gray.Print(idPadded)
		gray.Print(" │ ")
		blue.Print(namePadded)
		gray.Print(" │ ")
		statusColor.Print(statePadded)
		gray.Print("│ ")
		fmt.Println(imagePadded)

		// Ports
		ports := formatPorts(c.Ports)
		if ports != "" {
			gray.Printf("  ↪ Ports: %s\n", ports)
		}

		// Status/uptime
		status := c.Status
		gray.Printf("  ⏱ %s\n", status)

		fmt.Println()
	}

	// Summary
	runningCount := 0
	for _, c := range containers {
		if c.State == "running" {
			runningCount++
		}
	}
	fmt.Printf("Total: %d containers", len(containers))
	if runningCount > 0 {
		green.Printf(" (%d running)", runningCount)
	}
	fmt.Println()
}

func formatPorts(ports []container.Port) string {
	if len(ports) == 0 {
		return ""
	}

	var portStrs []string
	for _, port := range ports {
		if port.PublicPort > 0 {
			portStrs = append(portStrs, fmt.Sprintf("%d:%d/%s", port.PublicPort, port.PrivatePort, port.Type))
		} else {
			portStrs = append(portStrs, fmt.Sprintf("%d/%s", port.PrivatePort, port.Type))
		}
	}

	result := strings.Join(portStrs, ", ")
	return result
}
