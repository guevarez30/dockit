package ui

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/guevarez30/dockit/docker"
)

// LogsModel represents the logs viewer
type LogsModel struct {
	client       *docker.Client
	containerID  string
	viewport     viewport.Model
	logs         []string
	filteredLogs []string
	exit         bool
	err          error
	keys         KeyMap
	ready        bool
	searchMode   bool
	searchInput  textinput.Model
	searchTerm   string
}

// NewLogsModel creates a new logs model
func NewLogsModel(client *docker.Client, containerID string) *LogsModel {
	ti := textinput.New()
	ti.Placeholder = "Search logs..."
	ti.CharLimit = 50

	return &LogsModel{
		client:      client,
		containerID: containerID,
		keys:        DefaultKeyMap(),
		viewport:    viewport.New(80, 20),
		searchInput: ti,
	}
}

// logsMsg is sent when logs are received
type logsMsg []string

// Init initializes the logs viewer
func (m *LogsModel) Init() tea.Cmd {
	return m.fetchLogs()
}

// Update handles messages
func (m *LogsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle search mode separately
		if m.searchMode {
			switch msg.String() {
			case "enter":
				// Apply search
				m.searchTerm = m.searchInput.Value()
				m.searchMode = false
				m.filterLogs()
				return m, nil
			case "esc":
				// Cancel search
				m.searchMode = false
				m.searchInput.SetValue("")
				return m, nil
			default:
				// Update search input
				m.searchInput, cmd = m.searchInput.Update(msg)
				return m, cmd
			}
		}

		// Normal mode key handling
		switch {
		case key.Matches(msg, m.keys.Back):
			// Clear search if active, otherwise exit
			if m.searchTerm != "" {
				m.searchTerm = ""
				m.searchInput.SetValue("")
				m.filterLogs()
			} else {
				m.exit = true
			}
			return m, nil
		case key.Matches(msg, m.keys.Search):
			// Enter search mode
			m.searchMode = true
			m.searchInput.Focus()
			return m, textinput.Blink
		case key.Matches(msg, m.keys.Up):
			m.viewport.LineUp(1)
		case key.Matches(msg, m.keys.Down):
			m.viewport.LineDown(1)
		}

	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, msg.Height-10)
			m.viewport.YPosition = 3
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - 10
		}

	case logsMsg:
		m.logs = msg
		m.filteredLogs = msg
		m.ready = true
		m.viewport.SetContent(strings.Join(m.logs, "\n"))
		m.viewport.GotoBottom()
		return m, nil

	case errMsg:
		m.err = msg
		return m, nil
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the logs viewer
func (m *LogsModel) View() string {
	if m.err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	title := TitleStyle.Render(fmt.Sprintf("Logs - %s", m.containerID[:12]))

	var content string
	if !m.ready {
		content = "Loading logs..."
	} else {
		content = m.viewport.View()
	}

	// Search bar
	var searchBar string
	if m.searchMode {
		searchBar = lipgloss.NewStyle().
			Foreground(infoColor).
			Render("Search: ") + m.searchInput.View()
	} else if m.searchTerm != "" {
		matchCount := len(m.filteredLogs)
		searchBar = lipgloss.NewStyle().
			Foreground(successColor).
			Render(fmt.Sprintf("Filtered: %d matches for '%s' (esc to clear)", matchCount, m.searchTerm))
	}

	// Help text
	var help string
	if m.searchMode {
		help = HelpStyle.Render("enter: apply • esc: cancel")
	} else if m.searchTerm != "" {
		help = HelpStyle.Render("↑/↓: scroll • /: new search • esc: clear filter • esc esc: back")
	} else {
		help = HelpStyle.Render("↑/↓: scroll • /: search • esc: back")
	}

	var parts []string
	parts = append(parts, title)
	if searchBar != "" {
		parts = append(parts, "", searchBar)
	}
	parts = append(parts, "", content, "", help)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// fetchLogs fetches container logs
func (m *LogsModel) fetchLogs() tea.Cmd {
	return func() tea.Msg {
		logReader, err := m.client.GetContainerLogs(m.containerID, false)
		if err != nil {
			return errMsg(err)
		}
		defer logReader.Close()

		var logs []string

		// Read all bytes
		data, err := io.ReadAll(logReader)
		if err != nil {
			return errMsg(err)
		}

		// If no logs, return empty
		if len(data) == 0 {
			return logsMsg([]string{"No logs available"})
		}

		// Docker uses a special header format for logs
		// Parse the docker log format (8 byte header per line)
		i := 0
		for i < len(data) {
			// Check if we have at least 8 bytes for header
			if i+8 > len(data) {
				break
			}

			// Skip the 8-byte header
			// Bytes 4-7 contain the size of the log line
			size := int(data[i+4])<<24 | int(data[i+5])<<16 | int(data[i+6])<<8 | int(data[i+7])
			i += 8

			// Extract the log line
			if i+size <= len(data) {
				line := string(data[i : i+size])
				line = strings.TrimSpace(line)
				if line != "" {
					logs = append(logs, line)
				}
				i += size
			} else {
				// If size is invalid, treat rest as one line
				line := string(data[i:])
				line = strings.TrimSpace(line)
				if line != "" {
					logs = append(logs, line)
				}
				break
			}
		}

		// If parsing failed, try simple line-by-line
		if len(logs) == 0 {
			scanner := bufio.NewScanner(strings.NewReader(string(data)))
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					logs = append(logs, line)
				}
			}
		}

		// If still no logs
		if len(logs) == 0 {
			logs = []string{"No logs available"}
		}

		// Limit to last 500 lines
		if len(logs) > 500 {
			logs = logs[len(logs)-500:]
		}

		return logsMsg(logs)
	}
}

// filterLogs filters the logs based on the search term
func (m *LogsModel) filterLogs() {
	if m.searchTerm == "" {
		m.filteredLogs = m.logs
		m.viewport.SetContent(strings.Join(m.logs, "\n"))
		return
	}

	// Case-insensitive search
	searchLower := strings.ToLower(m.searchTerm)
	var filtered []string

	for _, line := range m.logs {
		if strings.Contains(strings.ToLower(line), searchLower) {
			// Highlight the match
			highlighted := highlightMatch(line, m.searchTerm)
			filtered = append(filtered, highlighted)
		}
	}

	if len(filtered) == 0 {
		filtered = []string{fmt.Sprintf("No matches found for '%s'", m.searchTerm)}
	}

	m.filteredLogs = filtered
	m.viewport.SetContent(strings.Join(filtered, "\n"))
	m.viewport.GotoTop()
}

// highlightMatch highlights search matches in the log line
func highlightMatch(line, term string) string {
	if term == "" {
		return line
	}

	termLower := strings.ToLower(term)
	lineLower := strings.ToLower(line)

	// Find all occurrences
	result := ""
	lastIndex := 0

	for {
		index := strings.Index(lineLower[lastIndex:], termLower)
		if index == -1 {
			result += line[lastIndex:]
			break
		}

		actualIndex := lastIndex + index
		result += line[lastIndex:actualIndex]

		// Highlight the match
		matchStyle := lipgloss.NewStyle().
			Background(warningColor).
			Foreground(lipgloss.Color("#000000")).
			Bold(true)
		result += matchStyle.Render(line[actualIndex : actualIndex+len(term)])

		lastIndex = actualIndex + len(term)
	}

	return result
}
