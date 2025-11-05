# Dockit - Architecture & Design Decisions

## Architecture Overview

Dockit follows a layered architecture with clear separation between Docker operations, business logic, and presentation.

```
┌─────────────────────────────────────────┐
│         User Interface Layer            │
│    (Bubble Tea Models & Views)          │
├─────────────────────────────────────────┤
│         Business Logic Layer            │
│   (State Management, Actions)           │
├─────────────────────────────────────────┤
│         Docker Client Layer             │
│    (Simplified API Wrapper)             │
├─────────────────────────────────────────┤
│         Docker Engine API               │
└─────────────────────────────────────────┘
```

## Layer Details

### 1. Docker Client Layer (`docker/client.go`)

**Purpose**: Simplify and wrap Docker Engine API calls

**Key Design Decisions**:
- Single `Client` struct wrapping official Docker SDK
- Context-based operations for cancellation support
- Simplified method signatures hiding API complexity
- Error handling at this layer

**Example Methods**:
```go
ListContainers(all bool) ([]types.Container, error)
StartContainer(id string) error
GetContainerLogs(id string, follow bool) (io.ReadCloser, error)
```

**Why**: Keeps UI layer clean and allows easy testing/mocking

---

### 2. UI Layer (`ui/`)

**Architecture Pattern**: The Elm Architecture (TEA) via Bubble Tea

#### Main Model (`ui/model.go`)

**Responsibilities**:
- Top-level application state
- View routing (Dashboard, Containers, Images, Logs)
- Tab navigation
- Window size management

**Key Components**:
```go
type Model struct {
    client      *docker.Client    // Docker operations
    currentView View              // Active view
    containers  *ContainersModel  // Sub-model
    images      *ImagesModel      // Sub-model
    dashboard   *DashboardModel   // Sub-model
    logs        *LogsModel        // Sub-model
}
```

**Message Flow**:
1. User input → KeyMsg
2. Model.Update() routes to active sub-model
3. Sub-model returns commands
4. Commands execute async → send new messages
5. Model.View() renders current state

#### Sub-Models Pattern

Each view (Dashboard, Containers, Images, Logs) implements:
- **Model struct** - View-specific state
- **Init() tea.Cmd** - Initialize and fetch data
- **Update(msg) (tea.Model, tea.Cmd)** - Handle messages
- **View() string** - Render UI
- **refresh() tea.Cmd** - Reload data

**Why**: Modularity, testability, clear ownership of state

---

## Key Design Decisions

### 1. Tab System Design

**Decision**: Tabs always visible on main views, hidden in logs

**Rationale**:
- Tabs provide clear navigation context
- Logs get full-screen for better readability
- Consistent UX across Dashboard/Containers/Images

**Implementation**:
```go
// model.go View()
switch m.currentView {
case LogsView:
    return m.logs.View()  // No tabs/footer
default:
    return tabs + content + footer
}
```

---

### 2. Container View Layout

**Decision**: Tabular format with fixed-width columns

**Columns**:
- STATUS (12 chars) - Visual indicator + state
- NAME (25 chars) - Container name
- IMAGE (30 chars) - Image name
- ID (12 chars) - Short container ID
- UPTIME (15 chars) - Time since started

**Why**:
- Predictable, clean layout
- Easy to scan
- Prevents text wrapping issues
- Truncation with "..." for overflow

**Evolution**: Originally had wrapping text issues → fixed with column widths

---

### 3. Action Feedback System

**Decision**: Visual indicators for async operations

**Implementation**:
```go
type ContainersModel struct {
    statusMsg        string  // Success message
    actionInProgress bool    // Loading state
}
```

**Flow**:
1. User presses `r` (restart)
2. `actionInProgress = true` → shows "⟳ Processing..."
3. Docker API call executes
4. On completion → `statusMsg = "✓ Container restarted"`
5. Auto-clear after 2 seconds
6. List refreshes

**Why**: User needs confirmation that actions are happening and completed

---

### 4. Log Viewer Design

**Challenge**: Docker logs use 8-byte header format per line

**Solution**: Custom parser in `fetchLogs()`
```go
// Parse 8-byte header (bytes 4-7 = size)
size := int(data[i+4])<<24 | int(data[i+5])<<16 | ...
// Extract log line
line := string(data[i+8 : i+8+size])
```

**Fallback**: If parsing fails, use simple line-by-line scanning

**Why**: Docker's wire format is complex; need robust parsing with fallback

---

### 5. Search Implementation

**Decision**: In-memory filtering with highlighting

**Features**:
- Case-insensitive search
- Real-time highlighting with background color
- Shows match count
- Preserves original logs

**Implementation**:
```go
type LogsModel struct {
    logs         []string  // Original logs
    filteredLogs []string  // Search results
    searchTerm   string    // Active search
}
```

**Why**: Fast, simple, works offline, no external dependencies

---

### 6. Uptime Calculation

**Decision**: Show relative time, not absolute timestamps

**Format**:
- `30s ago` - Under 1 minute
- `5m ago` - Under 1 hour
- `2h ago` - Under 24 hours
- `3d ago` - Over 24 hours

**Why**: More useful for monitoring restart frequency

**Calculation**: Based on container's `Created` timestamp (Unix seconds)

---

## Styling System (`ui/styles.go`)

### Color Palette
```go
primaryColor   = "#7D56F4"  // Purple - active elements
secondaryColor = "#FF79C6"  // Pink - accents
successColor   = "#50FA7B"  // Green - running, success
warningColor   = "#FFB86C"  // Orange - warnings, highlights
errorColor     = "#FF5555"  // Red - errors, stopped
infoColor      = "#8BE9FD"  // Cyan - labels, info
mutedColor     = "#6272A4"  // Gray - inactive, help text
```

### Style Patterns
- **ActiveTabStyle**: Bold, white text, purple background
- **InactiveTabStyle**: Muted gray text, no background
- **CardStyle**: Rounded borders, padding, margin
- **StatusStyles**: Color-coded by state (running/stopped/paused)

### Border Usage
- Tabs: No borders (cleaner look after iteration)
- Cards: Rounded borders for dashboard stats
- Title: Bottom border only

---

## Message Patterns

### Custom Message Types
```go
// Data messages
type containersMsg []types.Container
type imagesMsg []image.Summary
type logsMsg []string

// Action messages
type containerActionMsg struct {
    success bool
    message string
}

// UI messages
type clearStatusMsg struct{}
```

### Async Patterns
```go
// Command that returns message
func (m *Model) refresh() tea.Cmd {
    return func() tea.Msg {
        data, err := m.client.ListContainers(true)
        if err != nil {
            return errMsg(err)
        }
        return containersMsg(data)
    }
}
```

---

## Error Handling

### Strategy
1. **API Layer**: Return Go errors
2. **Model Layer**: Convert to `errMsg` type
3. **View Layer**: Display with `ErrorStyle`

### User Experience
- Errors shown in red box
- Operations continue (no crashes)
- Error cleared on next successful action

---

## Performance Considerations

### Current
- Sync operations (blocking UI)
- Full refresh on actions
- In-memory log storage (500 line limit)

### Future Improvements
- Background polling for container states
- Incremental updates
- Streaming logs
- Virtual scrolling for large lists

---

## Testing Strategy (Future)

### Planned
- Unit tests for Docker client wrapper
- Mock Docker client for UI testing
- Snapshot tests for view rendering
- Integration tests with test containers

### Challenges
- Bubble Tea is inherently stateful
- Terminal rendering hard to test
- Async commands need careful mocking
