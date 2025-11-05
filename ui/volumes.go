package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/api/types/volume"
	"github.com/guevarez30/dockit/docker"
)

// VolumesModel represents the volumes view
type VolumesModel struct {
	client           *docker.Client
	volumes          []*volume.Volume
	cursor           int
	err              error
	keys             KeyMap
	statusMsg        string
	actionInProgress bool
}

// NewVolumesModel creates a new volumes model
func NewVolumesModel(client *docker.Client) *VolumesModel {
	return &VolumesModel{
		client: client,
		keys:   DefaultKeyMap(),
	}
}

// volumesMsg is sent when volumes are loaded
type volumesMsg []*volume.Volume

// volumeActionMsg is sent after a volume action completes
type volumeActionMsg struct {
	success bool
	message string
}

// Init initializes the volumes view
func (m *VolumesModel) Init() tea.Cmd {
	return m.refresh()
}

// Update handles messages
func (m *VolumesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.volumes)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.Remove):
			return m, m.removeVolume()
		case key.Matches(msg, m.keys.Refresh):
			return m, m.refresh()
		}

	case volumesMsg:
		m.volumes = msg
		m.actionInProgress = false
		if m.cursor >= len(m.volumes) {
			m.cursor = len(m.volumes) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
		return m, nil

	case volumeActionMsg:
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

// View renders the volumes view
func (m *VolumesModel) View() string {
	if m.err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	if len(m.volumes) == 0 {
		return HelpStyle.Render("No volumes found")
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
		Render(fmt.Sprintf("%-30s  %-15s  %-15s  %-40s", "NAME", "DRIVER", "SCOPE", "MOUNTPOINT"))

	rows = append(rows, header)
	rows = append(rows, "") // Empty line after header

	for i, vol := range m.volumes {
		row := m.renderVolumeRow(vol, i == m.cursor)
		rows = append(rows, row)
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// renderVolumeRow renders a single volume row
func (m *VolumesModel) renderVolumeRow(vol *volume.Volume, selected bool) string {
	// Volume name
	name := vol.Name
	if len(name) > 30 {
		name = name[:27] + "..."
	}

	// Driver
	driver := vol.Driver
	if len(driver) > 15 {
		driver = driver[:12] + "..."
	}

	// Scope
	scope := vol.Scope
	if len(scope) > 15 {
		scope = scope[:12] + "..."
	}

	// Mountpoint
	mountpoint := vol.Mountpoint
	if len(mountpoint) > 40 {
		// Show start and end
		mountpoint = mountpoint[:18] + "..." + mountpoint[len(mountpoint)-19:]
	}

	row := fmt.Sprintf("%-30s  %-15s  %-15s  %-40s",
		name,
		driver,
		scope,
		mountpoint)

	if selected {
		return lipgloss.NewStyle().
			Background(primaryColor).
			Foreground(lipgloss.Color("#FAFAFA")).
			Padding(0, 1).
			Render(row)
	}

	return lipgloss.NewStyle().Padding(0, 1).Render(row)
}

// refresh fetches the latest volumes
func (m *VolumesModel) refresh() tea.Cmd {
	return func() tea.Msg {
		volumes, err := m.client.ListVolumes()
		if err != nil {
			return errMsg(err)
		}
		return volumesMsg(volumes)
	}
}

// removeVolume removes the selected volume
func (m *VolumesModel) removeVolume() tea.Cmd {
	if len(m.volumes) == 0 {
		return nil
	}

	m.actionInProgress = true
	vol := m.volumes[m.cursor]
	return func() tea.Msg {
		err := m.client.RemoveVolume(vol.Name, false)
		if err != nil {
			return errMsg(err)
		}
		return volumeActionMsg{success: true, message: "Volume removed"}
	}
}

// clearStatusAfter clears the status message after a duration
func (m *VolumesModel) clearStatusAfter(duration time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(duration)
		return clearStatusMsg{}
	}
}
