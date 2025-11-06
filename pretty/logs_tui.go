package pretty

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00d7ff")).
			MarginBottom(1)

	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#3a3a3a")).
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 1)

	searchBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#ffff00")).
			Foreground(lipgloss.Color("#000000")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	highlightStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#ffff00")).
			Foreground(lipgloss.Color("#000000")).
			Bold(true)
)

type logLine struct {
	raw       string
	formatted string
	timestamp time.Time
}

type logsModel struct {
	containerID   string
	containerName string
	lines         []logLine
	scrollOffset  int
	width         int
	height        int
	follow        bool
	paused        bool
	searchMode    bool
	searchInput   textinput.Model
	searchPattern *regexp.Regexp
	matchCount    int
	currentMatch  int
	reader        io.ReadCloser
	ctx           context.Context
	cancel        context.CancelFunc
	done          bool
}

type logMsg struct {
	line logLine
}

type errMsg struct {
	err error
}

func (m logsModel) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.readLogs(),
	)
}

func (m logsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searchMode {
			switch msg.String() {
			case "enter":
				// Apply search
				m.searchMode = false
				pattern := m.searchInput.Value()
				if pattern != "" {
					compiled, err := regexp.Compile("(?i)" + pattern)
					if err == nil {
						m.searchPattern = compiled
						m.updateMatchCount()
						m.currentMatch = 0
						m.jumpToNextMatch()
					}
				} else {
					m.searchPattern = nil
					m.matchCount = 0
				}
				return m, nil
			case "esc":
				m.searchMode = false
				m.searchInput.SetValue("")
				return m, nil
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				return m, cmd
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			m.cleanup()
			return m, tea.Quit
		case "/":
			m.searchMode = true
			m.searchInput.Focus()
			return m, nil
		case "n":
			if m.searchPattern != nil {
				m.jumpToNextMatch()
			}
			return m, nil
		case "N":
			if m.searchPattern != nil {
				m.jumpToPrevMatch()
			}
			return m, nil
		case " ":
			m.paused = !m.paused
			return m, nil
		case "up", "k":
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}
			return m, nil
		case "down", "j":
			maxScroll := max(0, len(m.lines)-m.contentHeight())
			if m.scrollOffset < maxScroll {
				m.scrollOffset++
			}
			return m, nil
		case "pgup":
			m.scrollOffset = max(0, m.scrollOffset-m.contentHeight())
			return m, nil
		case "pgdown":
			maxScroll := max(0, len(m.lines)-m.contentHeight())
			m.scrollOffset = min(m.scrollOffset+m.contentHeight(), maxScroll)
			return m, nil
		case "home", "g":
			m.scrollOffset = 0
			return m, nil
		case "end", "G":
			m.scrollOffset = max(0, len(m.lines)-m.contentHeight())
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case logMsg:
		if !m.paused {
			m.lines = append(m.lines, msg.line)
			// Auto-scroll to bottom if we're following and near the end
			if m.follow {
				maxScroll := max(0, len(m.lines)-m.contentHeight())
				if m.scrollOffset >= maxScroll-5 { // Within 5 lines of bottom
					m.scrollOffset = maxScroll
				}
			}
			m.updateMatchCount()
		}
		if !m.done {
			return m, m.readLogs()
		}
		return m, nil

	case errMsg:
		m.done = true
		return m, nil
	}

	return m, nil
}

func (m logsModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	var sb strings.Builder

	// Title
	title := titleStyle.Render(fmt.Sprintf("📋 LOGS: %s", m.containerName))
	sb.WriteString(title)
	sb.WriteString("\n")

	// Content area
	contentHeight := m.contentHeight()
	visibleLines := m.getVisibleLines(contentHeight)

	for _, line := range visibleLines {
		formatted := m.formatLine(line)
		sb.WriteString(formatted)
		sb.WriteString("\n")
	}

	// Pad remaining space
	for i := len(visibleLines); i < contentHeight; i++ {
		sb.WriteString("\n")
	}

	// Status bar
	statusBar := m.renderStatusBar()
	sb.WriteString(statusBar)

	// Search bar (if in search mode)
	if m.searchMode {
		sb.WriteString("\n")
		sb.WriteString(searchBarStyle.Render("Search: ") + m.searchInput.View())
	}

	return sb.String()
}

func (m *logsModel) contentHeight() int {
	// Title (2 lines with margin), status bar (1 line), search bar (1 line if active)
	reserved := 3
	if m.searchMode {
		reserved++
	}
	return max(1, m.height-reserved)
}

func (m *logsModel) getVisibleLines(count int) []logLine {
	start := m.scrollOffset
	end := min(start+count, len(m.lines))

	if start >= len(m.lines) {
		return []logLine{}
	}

	return m.lines[start:end]
}

func (m *logsModel) formatLine(line logLine) string {
	text := line.raw

	// Skip the Docker header bytes if present
	if len(text) > 8 {
		text = text[8:]
	}

	// Apply search highlighting
	if m.searchPattern != nil {
		if !m.searchPattern.MatchString(text) {
			// Don't show non-matching lines when search is active
			return ""
		}
		text = m.highlightMatches(text)
	}

	// Return raw text, preserving original terminal colors
	return text
}

func (m *logsModel) highlightMatches(text string) string {
	matches := m.searchPattern.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return text
	}

	var result strings.Builder
	lastEnd := 0

	for _, match := range matches {
		start, end := match[0], match[1]
		result.WriteString(text[lastEnd:start])
		result.WriteString(highlightStyle.Render(text[start:end]))
		lastEnd = end
	}

	result.WriteString(text[lastEnd:])
	return result.String()
}

func (m *logsModel) renderStatusBar() string {
	pauseIndicator := ""
	if m.paused {
		pauseIndicator = " [PAUSED]"
	}

	followIndicator := ""
	if m.follow {
		followIndicator = " [FOLLOW]"
	}

	searchInfo := ""
	if m.searchPattern != nil {
		searchInfo = fmt.Sprintf(" | Matches: %d", m.matchCount)
	}

	status := fmt.Sprintf("Lines: %d/%d%s%s%s",
		m.scrollOffset+1,
		len(m.lines),
		pauseIndicator,
		followIndicator,
		searchInfo,
	)

	help := "q: quit | /: search | n/N: next/prev | ↑↓: scroll | space: pause | g/G: top/bottom"

	// Calculate available width
	availWidth := m.width - lipgloss.Width(status) - 4

	if availWidth < len(help) {
		help = "q: quit | /: search | space: pause"
	}

	left := statusBarStyle.Render(status)
	right := statusBarStyle.Render(help)

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	return left + strings.Repeat(" ", gap) + right
}

func (m *logsModel) readLogs() tea.Cmd {
	return func() tea.Msg {
		if m.reader == nil {
			return errMsg{fmt.Errorf("reader is nil")}
		}

		scanner := bufio.NewScanner(m.reader)
		if scanner.Scan() {
			line := logLine{
				raw:       scanner.Text(),
				timestamp: time.Now(),
			}
			return logMsg{line: line}
		}

		if err := scanner.Err(); err != nil && err != io.EOF {
			return errMsg{err}
		}

		m.done = true
		return nil
	}
}

func (m *logsModel) updateMatchCount() {
	if m.searchPattern == nil {
		m.matchCount = 0
		return
	}

	count := 0
	for _, line := range m.lines {
		text := line.raw
		if len(text) > 8 {
			text = text[8:]
		}
		if m.searchPattern.MatchString(text) {
			count++
		}
	}
	m.matchCount = count
}

func (m *logsModel) jumpToNextMatch() {
	if m.searchPattern == nil || m.matchCount == 0 {
		return
	}

	for i := m.scrollOffset + 1; i < len(m.lines); i++ {
		text := m.lines[i].raw
		if len(text) > 8 {
			text = text[8:]
		}
		if m.searchPattern.MatchString(text) {
			m.scrollOffset = i
			return
		}
	}

	// Wrap around to beginning
	for i := 0; i <= m.scrollOffset; i++ {
		text := m.lines[i].raw
		if len(text) > 8 {
			text = text[8:]
		}
		if m.searchPattern.MatchString(text) {
			m.scrollOffset = i
			return
		}
	}
}

func (m *logsModel) jumpToPrevMatch() {
	if m.searchPattern == nil || m.matchCount == 0 {
		return
	}

	for i := m.scrollOffset - 1; i >= 0; i-- {
		text := m.lines[i].raw
		if len(text) > 8 {
			text = text[8:]
		}
		if m.searchPattern.MatchString(text) {
			m.scrollOffset = i
			return
		}
	}

	// Wrap around to end
	for i := len(m.lines) - 1; i >= m.scrollOffset; i-- {
		text := m.lines[i].raw
		if len(text) > 8 {
			text = text[8:]
		}
		if m.searchPattern.MatchString(text) {
			m.scrollOffset = i
			return
		}
	}
}

func (m *logsModel) cleanup() {
	if m.cancel != nil {
		m.cancel()
	}
	if m.reader != nil {
		m.reader.Close()
	}
}

// LaunchLogsTUI starts the TUI for viewing container logs
func LaunchLogsTUI(containerID string, follow bool) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("error creating Docker client: %v", err)
	}
	defer cli.Close()

	ctx, cancel := context.WithCancel(context.Background())

	// Get container info
	containerInfo, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		cancel()
		return fmt.Errorf("error inspecting container: %v", err)
	}

	// Get logs
	logOptions := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Timestamps: false,
		Tail:       "100", // Start with last 100 lines
	}

	reader, err := cli.ContainerLogs(ctx, containerID, logOptions)
	if err != nil {
		cancel()
		return fmt.Errorf("error getting container logs: %v", err)
	}

	// Initialize search input
	ti := textinput.New()
	ti.Placeholder = "Enter search pattern (regex supported)"
	ti.CharLimit = 100
	ti.Width = 50

	model := logsModel{
		containerID:   containerID,
		containerName: containerInfo.Name[1:], // Remove leading /
		lines:         []logLine{},
		follow:        follow,
		reader:        reader,
		ctx:           ctx,
		cancel:        cancel,
		searchInput:   ti,
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		cancel()
		reader.Close()
		return fmt.Errorf("error running TUI: %v", err)
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
