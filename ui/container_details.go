package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/guevarez30/dockit/docker"
)

// ContainerDetailsModel represents the container details view
type ContainerDetailsModel struct {
	client      *docker.Client
	containerID string
	inspect     types.ContainerJSON
	stats       *container.Stats
	err         error
	keys        KeyMap
	exit        bool
	scrollOffset int
}

// NewContainerDetailsModel creates a new container details model
func NewContainerDetailsModel(client *docker.Client, containerID string) *ContainerDetailsModel {
	return &ContainerDetailsModel{
		client:      client,
		containerID: containerID,
		keys:        DefaultKeyMap(),
	}
}

// containerDetailsMsg is sent when container details are loaded
type containerDetailsMsg struct {
	inspect types.ContainerJSON
	stats   *container.Stats
}

// Init initializes the container details view
func (m *ContainerDetailsModel) Init() tea.Cmd {
	return m.loadDetails()
}

// Update handles messages
func (m *ContainerDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Back):
			m.exit = true
			return m, nil
		case key.Matches(msg, m.keys.Up):
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}
		case key.Matches(msg, m.keys.Down):
			m.scrollOffset++
		case key.Matches(msg, m.keys.Refresh):
			return m, m.loadDetails()
		}

	case containerDetailsMsg:
		m.inspect = msg.inspect
		m.stats = msg.stats
		return m, nil

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, nil
}

// View renders the container details view
func (m *ContainerDetailsModel) View() string {
	// Safety checks
	if m == nil {
		return "Error: model is nil"
	}

	// Check for errors first
	if m.err != nil {
		footer := "Press ESC to go back"
		errMsg := fmt.Sprintf("Error: %v\n\n%s", m.err, footer)
		return lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true).
			Padding(1, 2).
			Render(errMsg)
	}

	// Check if data is loaded
	if m.inspect.ID == "" || m.inspect.Config == nil {
		loadingMsg := "Loading container details..."
		return lipgloss.NewStyle().
			Foreground(mutedColor).
			Padding(1, 2).
			Render(loadingMsg)
	}

	var sections []string

	// Title
	containerName := m.getContainerName()
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Padding(1, 2).
		Render(fmt.Sprintf("Container Details: %s", containerName))
	sections = append(sections, title)

	// Stats section
	sections = append(sections, m.renderStats())

	// Environment section
	sections = append(sections, m.renderEnvironment())

	// Configuration section
	sections = append(sections, m.renderConfiguration())

	// Footer
	footer := lipgloss.NewStyle().
		Foreground(mutedColor).
		Padding(1, 2).
		Render("↑/↓: scroll • r: refresh • esc: back")

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Apply scrolling
	lines := strings.Split(content, "\n")
	if m.scrollOffset > len(lines)-20 {
		m.scrollOffset = len(lines) - 20
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}
	}

	visibleLines := lines
	if m.scrollOffset < len(lines) {
		end := m.scrollOffset + 30
		if end > len(lines) {
			end = len(lines)
		}
		visibleLines = lines[m.scrollOffset:end]
	}

	return lipgloss.JoinVertical(lipgloss.Left, strings.Join(visibleLines, "\n"), "", footer)
}

// renderStats renders the statistics section
func (m *ContainerDetailsModel) renderStats() string {
	sectionTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(infoColor).
		Padding(0, 2).
		Render("STATISTICS")

	// Safety checks
	if m == nil || m.stats == nil || m.inspect.State == nil || !m.inspect.State.Running {
		content := lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(mutedColor).
			Render("Container is not running")
		return lipgloss.JoinVertical(lipgloss.Left, sectionTitle, content, "")
	}

	// Use the already-parsed stats
	statsData := m.stats

	// Calculate CPU percentage
	cpuPercent := calculateCPUPercent(statsData)

	// Calculate memory usage
	var memUsage, memLimit, memPercent float64
	if statsData.MemoryStats.Limit > 0 {
		memUsage = float64(statsData.MemoryStats.Usage) / 1024 / 1024
		memLimit = float64(statsData.MemoryStats.Limit) / 1024 / 1024
		memPercent = (float64(statsData.MemoryStats.Usage) / float64(statsData.MemoryStats.Limit)) * 100
	}

	// Network I/O
	var netRx, netTx uint64
	for _, netStats := range statsData.Networks {
		netRx += netStats.RxBytes
		netTx += netStats.TxBytes
	}

	// Block I/O
	var blkRead, blkWrite uint64
	for _, blkStat := range statsData.BlkioStats.IoServiceBytesRecursive {
		if blkStat.Op == "read" || blkStat.Op == "Read" {
			blkRead += blkStat.Value
		} else if blkStat.Op == "write" || blkStat.Op == "Write" {
			blkWrite += blkStat.Value
		}
	}

	statsContent := lipgloss.NewStyle().
		Padding(0, 4).
		Render(fmt.Sprintf(
			"CPU:         %.2f%%\n"+
				"Memory:      %.2f MiB / %.2f MiB (%.2f%%)\n"+
				"Network I/O: %s / %s\n"+
				"Block I/O:   %s / %s",
			cpuPercent,
			memUsage, memLimit, memPercent,
			formatBytes(netRx), formatBytes(netTx),
			formatBytes(blkRead), formatBytes(blkWrite),
		))

	return lipgloss.JoinVertical(lipgloss.Left, sectionTitle, statsContent, "")
}

// renderEnvironment renders the environment variables section
func (m *ContainerDetailsModel) renderEnvironment() string {
	sectionTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(infoColor).
		Padding(0, 2).
		Render("ENVIRONMENT VARIABLES")

	// Safety checks
	if m == nil || m.inspect.Config == nil || len(m.inspect.Config.Env) == 0 {
		content := lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(mutedColor).
			Render("No environment variables")
		return lipgloss.JoinVertical(lipgloss.Left, sectionTitle, content, "")
	}

	var envLines []string
	for _, env := range m.inspect.Config.Env {
		envLines = append(envLines, "  "+env)
	}

	envContent := lipgloss.NewStyle().
		Padding(0, 2).
		Render(strings.Join(envLines, "\n"))

	return lipgloss.JoinVertical(lipgloss.Left, sectionTitle, envContent, "")
}

// renderConfiguration renders the configuration section
func (m *ContainerDetailsModel) renderConfiguration() string {
	sectionTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(infoColor).
		Padding(0, 2).
		Render("CONFIGURATION")

	// Safety checks
	if m == nil || m.inspect.Config == nil {
		content := lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(mutedColor).
			Render("No configuration available")
		return lipgloss.JoinVertical(lipgloss.Left, sectionTitle, content, "")
	}

	var configLines []string

	// Image
	configLines = append(configLines, fmt.Sprintf("Image:       %s", m.inspect.Config.Image))

	// Status
	status := "Stopped"
	if m.inspect.State != nil && m.inspect.State.Running {
		status = "Running"
	}
	configLines = append(configLines, fmt.Sprintf("Status:      %s", status))

	// Command
	if len(m.inspect.Config.Cmd) > 0 {
		configLines = append(configLines, fmt.Sprintf("Command:     %s", strings.Join(m.inspect.Config.Cmd, " ")))
	}

	// Ports
	if len(m.inspect.Config.ExposedPorts) > 0 {
		var ports []string
		for port := range m.inspect.Config.ExposedPorts {
			ports = append(ports, string(port))
		}
		configLines = append(configLines, fmt.Sprintf("Ports:       %s", strings.Join(ports, ", ")))
	}

	// Port bindings
	if m.inspect.HostConfig != nil && len(m.inspect.HostConfig.PortBindings) > 0 {
		configLines = append(configLines, "Port Bindings:")
		for containerPort, hostBindings := range m.inspect.HostConfig.PortBindings {
			for _, binding := range hostBindings {
				configLines = append(configLines, fmt.Sprintf("  %s -> %s:%s", containerPort, binding.HostIP, binding.HostPort))
			}
		}
	}

	// Volumes/Mounts
	if len(m.inspect.Mounts) > 0 {
		configLines = append(configLines, "Mounts:")
		for _, mount := range m.inspect.Mounts {
			mountType := string(mount.Type)
			configLines = append(configLines, fmt.Sprintf("  [%s] %s -> %s", mountType, mount.Source, mount.Destination))
		}
	}

	// Networks
	if m.inspect.NetworkSettings != nil && len(m.inspect.NetworkSettings.Networks) > 0 {
		configLines = append(configLines, "Networks:")
		for netName, netConfig := range m.inspect.NetworkSettings.Networks {
			configLines = append(configLines, fmt.Sprintf("  %s (IP: %s)", netName, netConfig.IPAddress))
		}
	}

	// Labels
	if len(m.inspect.Config.Labels) > 0 {
		configLines = append(configLines, "Labels:")
		for key, value := range m.inspect.Config.Labels {
			if len(value) > 60 {
				value = value[:57] + "..."
			}
			configLines = append(configLines, fmt.Sprintf("  %s=%s", key, value))
		}
	}

	// Restart policy
	if m.inspect.HostConfig != nil && m.inspect.HostConfig.RestartPolicy.Name != "" {
		configLines = append(configLines, fmt.Sprintf("Restart:     %s", m.inspect.HostConfig.RestartPolicy.Name))
	}

	configContent := lipgloss.NewStyle().
		Padding(0, 4).
		Render(strings.Join(configLines, "\n"))

	return lipgloss.JoinVertical(lipgloss.Left, sectionTitle, configContent, "")
}

// loadDetails loads the container details and stats
func (m *ContainerDetailsModel) loadDetails() tea.Cmd {
	return func() tea.Msg {
		inspect, err := m.client.InspectContainer(m.containerID)
		if err != nil {
			return errMsg(err)
		}

		var stats *container.Stats
		if inspect.State != nil && inspect.State.Running {
			statsResp, err := m.client.GetContainerStats(m.containerID)
			if err == nil {
				// Parse stats immediately
				statsJSON, err := io.ReadAll(statsResp.Body)
				statsResp.Body.Close()
				if err == nil {
					var parsedStats container.Stats
					if err := json.Unmarshal(statsJSON, &parsedStats); err == nil {
						stats = &parsedStats
					}
				}
			}
		}

		return containerDetailsMsg{
			inspect: inspect,
			stats:   stats,
		}
	}
}

// getContainerName returns a clean container name
func (m *ContainerDetailsModel) getContainerName() string {
	if m == nil {
		return "unknown"
	}
	if m.inspect.Name != "" {
		name := m.inspect.Name
		if strings.HasPrefix(name, "/") {
			name = name[1:]
		}
		return name
	}
	if len(m.containerID) >= 12 {
		return m.containerID[:12]
	}
	return m.containerID
}

// calculateCPUPercent calculates the CPU usage percentage
func calculateCPUPercent(stats *container.Stats) float64 {
	// Safety check for nil stats pointer
	if stats == nil {
		return 0.0
	}

	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		return (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}
	return 0.0
}

// formatBytes formats bytes into a human-readable string
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
