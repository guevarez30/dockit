package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/api/types"
	"github.com/guevarez30/dockit/docker"
)

// ContainersModel represents the containers view
type ContainersModel struct {
	client       *docker.Client
	containers   []types.Container
	cursor       int
	selectedID   string
	showingLogs  bool
	showingDetails bool
	err          error
	keys         KeyMap
	statusMsg    string
	actionInProgress bool
}

// NewContainersModel creates a new containers model
func NewContainersModel(client *docker.Client) *ContainersModel {
	return &ContainersModel{
		client: client,
		keys:   DefaultKeyMap(),
	}
}

// containersMsg is sent when containers are loaded
type containersMsg []types.Container

// containerActionMsg is sent after a container action completes
type containerActionMsg struct {
	success bool
	message string
}

// clearStatusMsg is sent to clear the status message
type clearStatusMsg struct{}

// Init initializes the containers view
func (m *ContainersModel) Init() tea.Cmd {
	return m.refresh()
}

// Update handles messages
func (m *ContainersModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.containers)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.Start):
			return m, m.startContainer()
		case key.Matches(msg, m.keys.Stop):
			return m, m.stopContainer()
		case key.Matches(msg, m.keys.Restart):
			return m, m.restartContainer()
		case key.Matches(msg, m.keys.Remove):
			return m, m.removeContainer()
		case key.Matches(msg, m.keys.Logs):
			if len(m.containers) > 0 {
				m.selectedID = m.containers[m.cursor].ID
				m.showingLogs = true
			}
		case key.Matches(msg, m.keys.Enter):
			if len(m.containers) > 0 {
				m.selectedID = m.containers[m.cursor].ID
				m.showingDetails = true
			}
		case key.Matches(msg, m.keys.Refresh):
			return m, m.refresh()
		}

	case containersMsg:
		m.containers = msg
		m.actionInProgress = false
		if m.cursor >= len(m.containers) {
			m.cursor = len(m.containers) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
		return m, nil

	case containerActionMsg:
		// Show success message and refresh
		m.statusMsg = msg.message
		m.actionInProgress = false
		return m, tea.Batch(
			m.refresh(),
			m.clearStatusAfter(2 * time.Second),
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

// View renders the containers view
func (m *ContainersModel) View() string {
	if m.err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	if len(m.containers) == 0 {
		return HelpStyle.Render("No containers found")
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
		Render(fmt.Sprintf("%-12s  %-25s  %-30s  %-12s  %-15s", "STATUS", "NAME", "IMAGE", "ID", "UPTIME"))

	rows = append(rows, header)
	rows = append(rows, "") // Empty line after header

	for i, container := range m.containers {
		row := m.renderContainerRow(container, i == m.cursor)
		rows = append(rows, row)
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// renderContainerRow renders a single container row
func (m *ContainersModel) renderContainerRow(container types.Container, selected bool) string {
	// Status indicator
	var statusStyle lipgloss.Style
	var statusIcon string
	switch container.State {
	case "running":
		statusStyle = RunningStyle
		statusIcon = "●"
	case "exited":
		statusStyle = StoppedStyle
		statusIcon = "●"
	case "paused":
		statusStyle = PausedStyle
		statusIcon = "●"
	default:
		statusStyle = lipgloss.NewStyle()
		statusIcon = "○"
	}

	// Container name (remove leading slash)
	name := container.Names[0]
	if strings.HasPrefix(name, "/") {
		name = name[1:]
	}
	if len(name) > 25 {
		name = name[:22] + "..."
	}

	// Container image - clean it up
	image := container.Image
	if len(image) > 30 {
		image = image[:27] + "..."
	}

	// Container ID (short)
	id := container.ID[:12]

	// Calculate uptime
	uptime := formatUptime(container.Created, container.State)

	// Build row with fixed widths
	statusText := statusStyle.Render(fmt.Sprintf("%s %-9s", statusIcon, container.State))

	row := fmt.Sprintf("%-12s  %-25s  %-30s  %-12s  %s",
		statusText,
		name,
		image,
		id,
		uptime)

	if selected {
		return lipgloss.NewStyle().
			Background(primaryColor).
			Foreground(lipgloss.Color("#FAFAFA")).
			Padding(0, 1).
			Render(row)
	}

	return lipgloss.NewStyle().Padding(0, 1).Render(row)
}

// refresh fetches the latest containers
func (m *ContainersModel) refresh() tea.Cmd {
	return func() tea.Msg {
		containers, err := m.client.ListContainers(true)
		if err != nil {
			return errMsg(err)
		}
		return containersMsg(containers)
	}
}

// startContainer starts the selected container
func (m *ContainersModel) startContainer() tea.Cmd {
	if len(m.containers) == 0 {
		return nil
	}

	m.actionInProgress = true
	container := m.containers[m.cursor]
	return func() tea.Msg {
		err := m.client.StartContainer(container.ID)
		if err != nil {
			return errMsg(err)
		}
		return containerActionMsg{success: true, message: "Container started"}
	}
}

// stopContainer stops the selected container
func (m *ContainersModel) stopContainer() tea.Cmd {
	if len(m.containers) == 0 {
		return nil
	}

	m.actionInProgress = true
	container := m.containers[m.cursor]
	return func() tea.Msg {
		err := m.client.StopContainer(container.ID)
		if err != nil {
			return errMsg(err)
		}
		return containerActionMsg{success: true, message: "Container stopped"}
	}
}

// restartContainer restarts the selected container
func (m *ContainersModel) restartContainer() tea.Cmd {
	if len(m.containers) == 0 {
		return nil
	}

	m.actionInProgress = true
	container := m.containers[m.cursor]
	return func() tea.Msg {
		err := m.client.RestartContainer(container.ID)
		if err != nil {
			return errMsg(err)
		}
		return containerActionMsg{success: true, message: "Container restarted"}
	}
}

// removeContainer removes the selected container
func (m *ContainersModel) removeContainer() tea.Cmd {
	if len(m.containers) == 0 {
		return nil
	}

	m.actionInProgress = true
	container := m.containers[m.cursor]
	return func() tea.Msg {
		err := m.client.RemoveContainer(container.ID, true)
		if err != nil {
			return errMsg(err)
		}
		return containerActionMsg{success: true, message: "Container removed"}
	}
}

// clearStatusAfter clears the status message after a duration
func (m *ContainersModel) clearStatusAfter(duration time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(duration)
		return clearStatusMsg{}
	}
}

// formatUptime formats the container uptime in a human-readable format
func formatUptime(created int64, state string) string {
	if state == "exited" {
		return "stopped"
	}

	createdTime := time.Unix(created, 0)
	duration := time.Since(createdTime)

	if duration < time.Minute {
		return fmt.Sprintf("%ds ago", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%dm ago", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(duration.Hours()))
	} else {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}
}
