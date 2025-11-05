# Dockit - Project Overview

## Project Summary
Dockit is a modern, interactive terminal UI (TUI) for managing Docker containers and images. Built with Go and the Bubble Tea framework, it provides a beautiful, keyboard-driven interface inspired by LazyDocker but with a fresh, modern aesthetic.

## Project Goals
- Create a fast, responsive Docker management tool
- Provide a modern, visually appealing terminal interface
- Support vim-style keybindings for power users
- Offer real-time container and image management
- Implement log viewing with search capabilities

## Technology Stack

### Core Frameworks
- **Go 1.24+** - Primary language
- **Bubble Tea** - TUI framework using The Elm Architecture (Model-View-Update)
- **Lipgloss** - Terminal styling library for colors and layouts
- **Bubbles** - Pre-built TUI components (viewport, textinput, keys)

### Docker Integration
- **Docker Engine SDK** (`github.com/docker/docker/client`) - Official Docker API client
- **Docker Compose Library** (`github.com/compose-spec/compose-go`) - For future Compose support

## Key Design Decisions

### Architecture Pattern
- **MVC-like structure** with Bubble Tea's Model-View-Update pattern
- Separation of concerns:
  - `docker/` - Docker API wrapper/client
  - `ui/` - Bubble Tea models and views
  - `components/` - Reusable UI components (future)

### Visual Design
- **Color Scheme**: Purple (#7D56F4), Pink (#FF79C6), Cyan (#8BE9FD)
- **Typography**: Clean, tabular layouts with proper column alignment
- **Status Colors**: Green (running), Red (stopped), Yellow (paused/warning)
- **Tab Navigation**: Active tab with purple background, inactive tabs muted

### User Experience
- Vim-style keybindings (hjkl navigation)
- Tab-based view switching
- Visual feedback for actions (status messages, loading indicators)
- Search functionality with highlighting
- Non-blocking operations with progress indicators

## Project Structure

```
dockit/
├── main.go              # Application entry point
├── docker/              # Docker client wrapper
│   └── client.go        # Simplified Docker API interface
├── ui/                  # Bubble Tea UI layer
│   ├── model.go         # Main application model & view routing
│   ├── dashboard.go     # Dashboard statistics view
│   ├── containers.go    # Container management view
│   ├── images.go        # Image management view
│   ├── logs.go          # Log viewer with search
│   ├── styles.go        # Lipgloss style definitions
│   └── keys.go          # Keybinding definitions
├── components/          # Reusable components (future)
├── .claude/             # Documentation for AI assistants
└── README.md            # User-facing documentation
```

## Current State (2025-11-05)

### Implemented Features
✅ Dashboard with container/image statistics
✅ Container list with status indicators
✅ Container operations (start, stop, restart, remove)
✅ Container uptime display
✅ Image list with size and tags
✅ Image removal
✅ Log viewer with real-time display
✅ Log search with highlighting
✅ Tab navigation between views
✅ Visual feedback for operations
✅ Vim-style keybindings

### Planned Features (Roadmap)
- [ ] Docker Compose support (up/down/logs)
- [ ] Volume management
- [ ] Network inspection
- [ ] Container stats (CPU, memory graphs)
- [ ] Fuzzy search/filtering
- [ ] Bulk operations
- [ ] Theme customization
- [ ] Configuration file support

## Development Context
This project was built collaboratively with Claude Code, with emphasis on:
- Modern Go patterns and idioms
- Clean, maintainable code structure
- User experience and visual polish
- Incremental feature development
- Real-time feedback and iteration
