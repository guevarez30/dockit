# Dockit 🐳

A modern, interactive terminal UI for managing Docker containers and images, built with Go and Bubble Tea.

## Features

- **Dashboard View**: Overview of your Docker environment with container and image statistics
- **Container Management**:
  - List all containers (running and stopped)
  - Start, stop, restart containers
  - Remove containers
  - View real-time logs
- **Image Management**:
  - List all images with size information
  - Remove images (including dangling images)
  - View image details
- **Volume Management**:
  - List all Docker volumes
  - View volume details (driver, scope, mountpoint)
  - Remove volumes
- **Network Management**:
  - List all Docker networks
  - View network details (driver, scope, ID)
  - Remove user-created networks
- **Modern TUI**: Beautiful, responsive interface with vim-style keybindings
- **Color-coded Status**: Easy identification of container states (running, stopped, paused)

## Installation

### Prerequisites
- Go 1.24 or higher
- Docker daemon running

### Build from Source

```bash
git clone https://github.com/guevarez30/dockit.git
cd dockit
go build -o dockit
```

## Usage

Simply run the compiled binary:

```bash
./dockit
```

### Keybindings

#### Global
- `tab` - Switch between views (Containers, Images, Volumes, Networks, Compose)
- `q` or `ctrl+c` - Quit application
- `?` - Show help

#### Navigation
- `↑` or `k` - Move up
- `↓` or `j` - Move down
- `←` or `h` - Move left
- `→` or `l` - Move right

#### Container View
- `s` - Start selected container
- `x` - Stop selected container
- `r` - Restart selected container
- `d` - Remove selected container
- `L` - View logs of selected container
- `ctrl+r` - Refresh container list

#### Images View
- `d` - Remove selected image
- `ctrl+r` - Refresh image list

#### Volumes View
- `d` - Remove selected volume
- `ctrl+r` - Refresh volume list

#### Networks View
- `d` - Remove selected network (system networks cannot be removed)
- `ctrl+r` - Refresh network list

#### Logs View
- `↑`/`↓` - Scroll through logs
- `esc` - Return to container view

## Architecture

```
dockit/
├── main.go           # Application entry point
├── docker/           # Docker client wrapper
│   └── client.go
├── ui/               # Bubble Tea UI components
│   ├── model.go      # Main application model
│   ├── dashboard.go  # Dashboard view
│   ├── containers.go # Containers view
│   ├── images.go     # Images view
│   ├── volumes.go    # Volumes view
│   ├── networks.go   # Networks view
│   ├── logs.go       # Logs viewer
│   ├── styles.go     # Lipgloss styles
│   └── keys.go       # Keybindings
└── components/       # Reusable components
```

## Technology Stack

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - Terminal UI framework using The Elm Architecture
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** - Style definitions and rendering
- **[Bubbles](https://github.com/charmbracelet/bubbles)** - Common TUI components
- **[Docker SDK](https://github.com/docker/docker)** - Official Docker Engine API client

## Design Philosophy

Dockit is inspired by LazyDocker but with a fresh, modern aesthetic:
- Clean, card-based layouts
- Vibrant color scheme (purple, pink, cyan)
- Smooth vim-style navigation
- Intuitive keyboard shortcuts
- Real-time updates

## Roadmap

- [ ] Docker Compose support
- [x] Volume management
- [x] Network management
- [ ] Container stats (CPU, memory)
- [ ] Fuzzy search/filtering
- [ ] Bulk operations
- [ ] Theme customization
- [ ] Configuration file support

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Acknowledgments

- Inspired by [LazyDocker](https://github.com/jesseduffield/lazydocker)
- Built with the amazing [Charm](https://charm.sh) ecosystem
