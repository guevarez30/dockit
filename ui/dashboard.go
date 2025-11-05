package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/guevarez30/dockit/docker"
)

// DashboardStats holds dashboard statistics
type DashboardStats struct {
	TotalContainers   int
	RunningContainers int
	StoppedContainers int
	TotalImages       int
	DanglingImages    int
}

// DashboardModel represents the dashboard view
type DashboardModel struct {
	client *docker.Client
	stats  DashboardStats
	err    error
}

// NewDashboardModel creates a new dashboard model
func NewDashboardModel(client *docker.Client) *DashboardModel {
	return &DashboardModel{
		client: client,
	}
}

// statsMsg is sent when stats are loaded
type statsMsg DashboardStats

// errMsg is sent when there's an error
type errMsg error

// Init initializes the dashboard
func (m *DashboardModel) Init() tea.Cmd {
	return m.refresh()
}

// Update handles messages
func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case statsMsg:
		m.stats = DashboardStats(msg)
		return m, nil
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, nil
}

// View renders the dashboard
func (m *DashboardModel) View() string {
	if m.err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	title := TitleStyle.Render("Docker Dashboard")

	// Container stats card
	containerCard := m.renderContainerCard()

	// Image stats card
	imageCard := m.renderImageCard()

	// Layout cards horizontally
	cards := lipgloss.JoinHorizontal(
		lipgloss.Top,
		containerCard,
		"  ",
		imageCard,
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		cards,
		"",
		HelpStyle.Render("Press tab to navigate between views"),
	)
}

// renderContainerCard renders the container statistics card
func (m *DashboardModel) renderContainerCard() string {
	content := fmt.Sprintf(
		"%s %d\n\n%s %d\n%s %d",
		LabelStyle.Render("Total Containers:"),
		m.stats.TotalContainers,
		RunningStyle.Render("● Running:"),
		m.stats.RunningContainers,
		StoppedStyle.Render("● Stopped:"),
		m.stats.StoppedContainers,
	)

	return CardStyle.Width(35).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render("🐳 Containers"),
			"",
			content,
		),
	)
}

// renderImageCard renders the image statistics card
func (m *DashboardModel) renderImageCard() string {
	content := fmt.Sprintf(
		"%s %d\n\n%s %d",
		LabelStyle.Render("Total Images:"),
		m.stats.TotalImages,
		LabelStyle.Render("Dangling:"),
		m.stats.DanglingImages,
	)

	return CardStyle.Width(35).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render("📦 Images"),
			"",
			content,
		),
	)
}

// refresh fetches the latest stats
func (m *DashboardModel) refresh() tea.Cmd {
	return func() tea.Msg {
		containers, err := m.client.ListContainers(true)
		if err != nil {
			return errMsg(err)
		}

		images, err := m.client.ListImages()
		if err != nil {
			return errMsg(err)
		}

		stats := DashboardStats{
			TotalContainers: len(containers),
			TotalImages:     len(images),
		}

		// Count running and stopped containers
		for _, c := range containers {
			if c.State == "running" {
				stats.RunningContainers++
			} else {
				stats.StoppedContainers++
			}
		}

		// Count dangling images
		for _, img := range images {
			if len(img.RepoTags) == 0 {
				stats.DanglingImages++
			}
		}

		return statsMsg(stats)
	}
}
