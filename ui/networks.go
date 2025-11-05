package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/api/types/network"
	"github.com/guevarez30/dockit/docker"
)

// NetworksModel represents the networks view
type NetworksModel struct {
	client           *docker.Client
	networks         []*network.Summary
	cursor           int
	err              error
	keys             KeyMap
	statusMsg        string
	actionInProgress bool
}

// NewNetworksModel creates a new networks model
func NewNetworksModel(client *docker.Client) *NetworksModel {
	return &NetworksModel{
		client: client,
		keys:   DefaultKeyMap(),
	}
}

// networksMsg is sent when networks are loaded
type networksMsg []*network.Summary

// networkActionMsg is sent after a network action completes
type networkActionMsg struct {
	success bool
	message string
}

// Init initializes the networks view
func (m *NetworksModel) Init() tea.Cmd {
	return m.refresh()
}

// Update handles messages
func (m *NetworksModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If there's an error, ESC dismisses it
		if m.err != nil && key.Matches(msg, m.keys.Back) {
			m.err = nil
			return m, nil
		}

		switch {
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.networks)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.Remove):
			return m, m.removeNetwork()
		case key.Matches(msg, m.keys.Refresh):
			return m, m.refresh()
		}

	case networksMsg:
		m.networks = msg
		m.actionInProgress = false
		if m.cursor >= len(m.networks) {
			m.cursor = len(m.networks) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
		return m, nil

	case networkActionMsg:
		m.statusMsg = msg.message
		m.actionInProgress = false
		return m, tea.Batch(
			m.refresh(),
			m.clearStatusAfter(2*time.Second),
		)

	case errMsg:
		m.err = msg
		m.actionInProgress = false
		return m, nil

	case clearStatusMsg:
		m.statusMsg = ""
		return m, nil
	}

	return m, nil
}

// View renders the networks view
func (m *NetworksModel) View() string {
	if m.err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	if len(m.networks) == 0 {
		return HelpStyle.Render("No networks found")
	}

	var rows []string

	// Status message if present
	if m.statusMsg != "" {
		statusStyle := lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true).
			Padding(0, 1)
		rows = append(rows, statusStyle.Render("✓ "+m.statusMsg))
		rows = append(rows, "")
	}

	// Action in progress indicator
	if m.actionInProgress {
		progressStyle := lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true).
			Padding(0, 1)
		rows = append(rows, progressStyle.Render("⟳ Processing..."))
		rows = append(rows, "")
	}

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(infoColor).
		Padding(0, 1).
		Render(fmt.Sprintf("%-25s  %-15s  %-12s  %-12s  %-15s", "NAME", "DRIVER", "SCOPE", "ID", "CREATED"))

	rows = append(rows, header)
	rows = append(rows, "") // Empty line after header

	for i, net := range m.networks {
		row := m.renderNetworkRow(net, i == m.cursor)
		rows = append(rows, row)
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// renderNetworkRow renders a single network row
func (m *NetworksModel) renderNetworkRow(net *network.Summary, selected bool) string {
	// Network name
	name := net.Name
	if len(name) > 25 {
		name = name[:22] + "..."
	}

	// Driver
	driver := net.Driver
	if len(driver) > 15 {
		driver = driver[:12] + "..."
	}

	// Scope
	scope := net.Scope
	if len(scope) > 12 {
		scope = scope[:9] + "..."
	}

	// Network ID (short)
	id := net.ID
	if len(id) > 12 {
		id = id[:12]
	}

	// Created time
	created := formatNetworkTime(net.Created)

	row := fmt.Sprintf("%-25s  %-15s  %-12s  %-12s  %-15s",
		name,
		driver,
		scope,
		id,
		created)

	if selected {
		return lipgloss.NewStyle().
			Background(primaryColor).
			Foreground(lipgloss.Color("#FAFAFA")).
			Padding(0, 1).
			Render(row)
	}

	return lipgloss.NewStyle().Padding(0, 1).Render(row)
}

// refresh fetches the latest networks
func (m *NetworksModel) refresh() tea.Cmd {
	return func() tea.Msg {
		networks, err := m.client.ListNetworks()
		if err != nil {
			return errMsg(err)
		}
		return networksMsg(networks)
	}
}

// removeNetwork removes the selected network
func (m *NetworksModel) removeNetwork() tea.Cmd {
	if len(m.networks) == 0 {
		return nil
	}

	m.actionInProgress = true
	net := m.networks[m.cursor]

	// Prevent removal of system networks
	if net.Name == "bridge" || net.Name == "host" || net.Name == "none" {
		return func() tea.Msg {
			return errMsg(fmt.Errorf("cannot remove system network: %s", net.Name))
		}
	}

	return func() tea.Msg {
		err := m.client.RemoveNetwork(net.ID)
		if err != nil {
			return errMsg(err)
		}
		return networkActionMsg{success: true, message: "Network removed"}
	}
}

// clearStatusAfter clears the status message after a duration
func (m *NetworksModel) clearStatusAfter(duration time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(duration)
		return clearStatusMsg{}
	}
}

// formatNetworkTime formats the network creation time
func formatNetworkTime(created time.Time) string {
	if created.IsZero() {
		return "N/A"
	}

	duration := time.Since(created)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		return fmt.Sprintf("%dm ago", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(duration.Hours()))
	} else {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}
}
