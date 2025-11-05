package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/api/types/image"
	"github.com/guevarez30/dockit/docker"
)

// ImagesModel represents the images view
type ImagesModel struct {
	client *docker.Client
	images []image.Summary
	cursor int
	err    error
	keys   KeyMap
}

// NewImagesModel creates a new images model
func NewImagesModel(client *docker.Client) *ImagesModel {
	return &ImagesModel{
		client: client,
		keys:   DefaultKeyMap(),
	}
}

// imagesMsg is sent when images are loaded
type imagesMsg []image.Summary

// imageActionMsg is sent after an image action completes
type imageActionMsg struct {
	success bool
	message string
}

// Init initializes the images view
func (m *ImagesModel) Init() tea.Cmd {
	return m.refresh()
}

// Update handles messages
func (m *ImagesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.images)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.Remove):
			return m, m.removeImage()
		case key.Matches(msg, m.keys.Refresh):
			return m, m.refresh()
		}

	case imagesMsg:
		m.images = msg
		if m.cursor >= len(m.images) {
			m.cursor = len(m.images) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
		return m, nil

	case imageActionMsg:
		// Refresh after action
		return m, m.refresh()

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, nil
}

// View renders the images view
func (m *ImagesModel) View() string {
	if m.err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	if len(m.images) == 0 {
		return HelpStyle.Render("No images found")
	}

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(infoColor).
		Padding(0, 1).
		Render(fmt.Sprintf("%-40s  %-12s  %-12s", "REPOSITORY:TAG", "IMAGE ID", "SIZE"))

	var rows []string
	rows = append(rows, header)
	rows = append(rows, "") // Empty line after header

	for i, img := range m.images {
		row := m.renderImageRow(img, i == m.cursor)
		rows = append(rows, row)
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// renderImageRow renders a single image row
func (m *ImagesModel) renderImageRow(img image.Summary, selected bool) string {
	// Repository and tag
	var repoTag string
	if len(img.RepoTags) > 0 {
		repoTag = img.RepoTags[0]
	} else {
		repoTag = "<none>:<none>"
	}

	// Truncate if too long
	if len(repoTag) > 40 {
		repoTag = repoTag[:37] + "..."
	}

	// Image ID (short)
	id := img.ID[7:19] // Remove "sha256:" prefix and show 12 chars

	// Size (convert to MB)
	sizeMB := float64(img.Size) / 1024 / 1024
	sizeStr := fmt.Sprintf("%.1f MB", sizeMB)

	// Dangling indicator
	danglingIndicator := ""
	if len(img.RepoTags) == 0 {
		danglingIndicator = lipgloss.NewStyle().Foreground(warningColor).Render(" [dangling]")
	}

	row := fmt.Sprintf("%-40s  %-12s  %-12s%s",
		repoTag,
		id,
		sizeStr,
		danglingIndicator)

	if selected {
		return lipgloss.NewStyle().
			Background(primaryColor).
			Foreground(lipgloss.Color("#FAFAFA")).
			Padding(0, 1).
			Render(row)
	}

	return lipgloss.NewStyle().Padding(0, 1).Render(row)
}

// refresh fetches the latest images
func (m *ImagesModel) refresh() tea.Cmd {
	return func() tea.Msg {
		images, err := m.client.ListImages()
		if err != nil {
			return errMsg(err)
		}
		return imagesMsg(images)
	}
}

// removeImage removes the selected image
func (m *ImagesModel) removeImage() tea.Cmd {
	if len(m.images) == 0 {
		return nil
	}

	img := m.images[m.cursor]
	return func() tea.Msg {
		err := m.client.RemoveImage(img.ID, true)
		if err != nil {
			return errMsg(err)
		}
		return imageActionMsg{success: true, message: "Image removed"}
	}
}

// formatRepoTag formats repository and tag for display
func formatRepoTag(tags []string) string {
	if len(tags) == 0 {
		return "<none>:<none>"
	}
	tag := tags[0]
	parts := strings.Split(tag, ":")
	if len(parts) == 2 {
		return tag
	}
	return tag + ":latest"
}
