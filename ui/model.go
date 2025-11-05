package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/guevarez30/dockit/docker"
)

// View represents different views in the application
type View int

const (
	ContainersView View = iota
	ImagesView
	VolumesView
	NetworksView
	LogsView
	ContainerDetailsView
)

// Model is the main application model
type Model struct {
	client          *docker.Client
	currentView     View
	width           int
	height          int
	keys            KeyMap
	err             error
	scrollOffset    int

	// Sub-models for different views
	containers      *ContainersModel
	images          *ImagesModel
	volumes         *VolumesModel
	networks        *NetworksModel
	logs            *LogsModel
	containerDetails *ContainerDetailsModel
	showingHelp     bool
}

// NewModel creates a new application model
func NewModel() (*Model, error) {
	client, err := docker.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &Model{
		client:      client,
		currentView: ContainersView,
		keys:        DefaultKeyMap(),
		containers:  NewContainersModel(client),
		images:      NewImagesModel(client),
		volumes:     NewVolumesModel(client),
		networks:    NewNetworksModel(client),
	}, nil
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.containers.refresh(),
		tea.EnterAltScreen,
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle help toggle
		if key.Matches(msg, m.keys.Help) {
			m.showingHelp = !m.showingHelp
			return m, nil
		}

		// If showing help, escape dismisses it
		if m.showingHelp && key.Matches(msg, m.keys.Back) {
			m.showingHelp = false
			return m, nil
		}

		// Don't process other keys when help is showing
		if m.showingHelp {
			return m, nil
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			m.client.Close()
			return m, tea.Quit

		case key.Matches(msg, m.keys.Tab):
			m.currentView = (m.currentView + 1) % 4
			switch m.currentView {
			case ContainersView:
				return m, m.containers.refresh()
			case ImagesView:
				return m, m.images.refresh()
			case VolumesView:
				return m, m.volumes.refresh()
			case NetworksView:
				return m, m.networks.refresh()
			}
			return m, nil

		case key.Matches(msg, m.keys.ShiftTab):
			m.currentView = (m.currentView - 1 + 4) % 4
			m.scrollOffset = 0 // Reset scroll when switching views
			switch m.currentView {
			case ContainersView:
				return m, m.containers.refresh()
			case ImagesView:
				return m, m.images.refresh()
			case VolumesView:
				return m, m.volumes.refresh()
			case NetworksView:
				return m, m.networks.refresh()
			}
			return m, nil

		case key.Matches(msg, m.keys.PageUp):
			// Scroll up
			m.scrollOffset -= 10
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
			}
			return m, nil

		case key.Matches(msg, m.keys.PageDown):
			// Scroll down
			m.scrollOffset += 10
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Handle j/k for viewport scrolling (process before views to enable scroll)
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "j":
			m.scrollOffset++
			// Don't return here, let the view also handle cursor movement
		case "k":
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}
			// Don't return here, let the view also handle cursor movement
		}
	}

	// Update current view
	switch m.currentView {
	case ContainersView:
		newContainers, containersCmd := m.containers.Update(msg)
		m.containers = newContainers.(*ContainersModel)
		cmd = containersCmd

		// Check if we need to switch to logs view
		if m.containers.showingLogs {
			m.currentView = LogsView
			m.logs = NewLogsModel(m.client, m.containers.selectedID)
			m.containers.showingLogs = false
			cmd = m.logs.Init()
		}

		// Check if we need to switch to details view
		if m.containers.showingDetails {
			m.currentView = ContainerDetailsView
			m.containerDetails = NewContainerDetailsModel(m.client, m.containers.selectedID)
			m.containers.showingDetails = false
			cmd = m.containerDetails.Init()
		}
	case ImagesView:
		newImages, imagesCmd := m.images.Update(msg)
		m.images = newImages.(*ImagesModel)
		cmd = imagesCmd
	case VolumesView:
		newVolumes, volumesCmd := m.volumes.Update(msg)
		m.volumes = newVolumes.(*VolumesModel)
		cmd = volumesCmd
	case NetworksView:
		newNetworks, networksCmd := m.networks.Update(msg)
		m.networks = newNetworks.(*NetworksModel)
		cmd = networksCmd
	case LogsView:
		if m.logs != nil {
			newLogs, logsCmd := m.logs.Update(msg)
			m.logs = newLogs.(*LogsModel)
			cmd = logsCmd

			// Check if we need to exit logs view
			if m.logs.exit {
				m.currentView = ContainersView
				m.containers.showingLogs = false
				cmd = m.containers.refresh()
			}
		}
	case ContainerDetailsView:
		if m.containerDetails != nil {
			newDetails, detailsCmd := m.containerDetails.Update(msg)
			m.containerDetails = newDetails.(*ContainerDetailsModel)
			cmd = detailsCmd

			// Check if we need to exit details view
			if m.containerDetails.exit {
				m.currentView = ContainersView
				m.containers.showingDetails = false
				cmd = m.containers.refresh()
			}
		}
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m Model) View() string {
	if m.err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	// For logs view, return full screen without tabs/footer
	if m.currentView == LogsView && m.logs != nil {
		return m.logs.View()
	}

	// For container details view, return full screen without tabs/footer
	if m.currentView == ContainerDetailsView && m.containerDetails != nil {
		return m.containerDetails.View()
	}

	// Render tabs (fixed header)
	tabs := m.renderTabs()

	// Add separator line under tabs
	separator := lipgloss.NewStyle().
		Foreground(mutedColor).
		Render(strings.Repeat("─", 100))

	// Render footer (fixed)
	footer := m.renderFooter()

	// Calculate available height for content
	// tabs (1 line) + separator (1 line) + empty line (1) + content + empty line (1) + footer (3 lines)
	headerHeight := 4  // tabs + separator + padding
	footerHeight := 4  // padding + footer
	availableHeight := m.height - headerHeight - footerHeight
	if availableHeight < 5 {
		availableHeight = 5
	}

	// If showing help, overlay the help content
	if m.showingHelp {
		helpOverlay := m.renderHelpOverlay()
		return tabs + "\n" + separator + "\n\n" + helpOverlay + "\n\n" + footer
	}

	// Render current view content
	var fullContent string
	switch m.currentView {
	case ContainersView:
		fullContent = m.containers.View()
	case ImagesView:
		fullContent = m.images.View()
	case VolumesView:
		fullContent = m.volumes.View()
	case NetworksView:
		fullContent = m.networks.View()
	}

	// Apply viewport to content (scrolling)
	visibleContent := m.applyViewport(fullContent, availableHeight)

	return tabs + "\n" + separator + "\n\n" + visibleContent + "\n\n" + footer
}

// applyViewport applies scrolling to content based on scrollOffset
func (m Model) applyViewport(content string, maxLines int) string {
	lines := strings.Split(content, "\n")

	// Adjust scroll offset if needed
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
	if m.scrollOffset > len(lines)-maxLines {
		m.scrollOffset = len(lines) - maxLines
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}
	}

	// Extract visible lines
	endLine := m.scrollOffset + maxLines
	if endLine > len(lines) {
		endLine = len(lines)
	}

	visibleLines := lines[m.scrollOffset:endLine]

	// Pad with empty lines if needed
	for len(visibleLines) < maxLines {
		visibleLines = append(visibleLines, "")
	}

	return strings.Join(visibleLines, "\n")
}

// renderTabs renders the navigation tabs
func (m Model) renderTabs() string {
	tabs := []string{}
	views := []struct {
		name string
		view View
	}{
		{"Containers", ContainersView},
		{"Images", ImagesView},
		{"Volumes", VolumesView},
		{"Networks", NetworksView},
	}

	for _, v := range views {
		if m.currentView == v.view {
			tabs = append(tabs, ActiveTabStyle.Render(v.name))
		} else {
			tabs = append(tabs, InactiveTabStyle.Render(v.name))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

// renderFooter renders the footer with help text
func (m Model) renderFooter() string {
	helpText := "tab: switch view • ↑/↓: navigate"

	switch m.currentView {
	case ContainersView:
		helpText += " • s: start • x: stop • r: restart • d: remove • L: logs • enter: details"
	case ImagesView:
		helpText += " • d: remove • enter: inspect"
	case VolumesView:
		helpText += " • d: remove"
	case NetworksView:
		helpText += " • d: remove"
	case LogsView:
		helpText += " • esc: back • ↑/↓: scroll"
	case ContainerDetailsView:
		helpText += " • esc: back • ↑/↓: scroll • r: refresh"
	}

	helpText += " • ?: help • q: quit"

	return FooterStyle.Render(helpText)
}

// renderHelpOverlay renders context-specific help as an overlay
func (m Model) renderHelpOverlay() string {
	var helpContent string

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Padding(0, 0, 1, 0).
		Render("HELP")

	commonHelp := `
NAVIGATION
  tab         Switch between views
  ↑/↓         Navigate lists
  esc         Close help
  q           Quit application
`

	switch m.currentView {
	case ContainersView:
		helpContent = `
CONTAINERS VIEW

View and manage your Docker containers. Running containers are displayed
with their current status, name, image, and ports.

COMMANDS
  s           Start selected container
  x           Stop selected container
  r           Restart selected container
  d           Remove selected container
  L           View container logs
  ↑/↓         Navigate container list
`
	case ImagesView:
		helpContent = `
IMAGES VIEW

Browse and manage Docker images on your system. View image names, tags,
sizes, and when they were created.

COMMANDS
  d           Remove selected image
  enter       Inspect image details
  ↑/↓         Navigate image list
`
	case VolumesView:
		helpContent = `
VOLUMES VIEW

Manage Docker volumes used for persistent data storage. View volume names,
drivers, and mount points.

COMMANDS
  d           Remove selected volume
  ↑/↓         Navigate volume list
`
	case NetworksView:
		helpContent = `
NETWORKS VIEW

View and manage Docker networks. See network names, drivers, and scopes.
System networks (bridge, host, none) cannot be removed.

COMMANDS
  d           Remove selected network
  ↑/↓         Navigate network list
`
	}

	footer := `
PROJECT
  GitHub: https://github.com/guevarez30/dockit
  For issues and contributions, visit the repository.
`

	helpBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		Render(title + helpContent + commonHelp + footer)

	return helpBox
}
