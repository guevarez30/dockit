# Dockit 🐳

A modern, interactive terminal UI for managing Docker containers, images, volumes, and networks, built with Go and Bubble Tea.

## ✨ Highlights

- 📊 **Live Container Stats** - Real-time CPU, memory, network, and disk I/O monitoring
- 🔍 **Deep Inspection** - View environment variables, ports, mounts, and network configs
- 🎨 **Beautiful UI** - Modern terminal interface with vim-style navigation
- ⚡ **Fast & Lightweight** - Built with Go for optimal performance
- 🎯 **Intuitive** - Context-aware help system and simple keyboard shortcuts

## Features

- **Dashboard View**: Overview of your Docker environment with container and image statistics
- **Container Management**:
  - List all containers (running and stopped)
  - Start, stop, restart containers
  - Remove containers
  - View real-time logs
  - **Detailed container inspection** with live stats (CPU, memory, network I/O, block I/O)
  - View environment variables, configuration, ports, volumes, and networks
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
- **Interactive Help System**: Context-aware help overlay accessible with `?` key

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

### Quick Start Examples

**View Container Stats**
1. Launch Dockit: `./dockit`
2. Navigate to a running container using `↑`/`↓`
3. Press `enter` to view detailed stats (CPU, memory, network I/O)
4. Press `r` to refresh stats in real-time
5. Press `esc` to return to containers view

**View Container Logs**
1. Select a container with `↑`/`↓`
2. Press `L` (Shift+l) to view logs
3. Scroll through logs with `↑`/`↓`
4. Press `esc` to return

**Start/Stop Containers**
1. Navigate to a stopped container
2. Press `s` to start or `x` to stop
3. Press `r` to restart a running container

**Remove Unused Resources**
1. Press `tab` to switch to Images/Volumes/Networks view
2. Navigate to unused resources
3. Press `d` to remove selected items

### Keybindings

#### Global
- `tab` / `shift+tab` - Switch between views (forward/backward)
- `q` or `ctrl+c` - Quit application
- `?` - Show context-aware help overlay
- `ctrl+r` - Refresh current view

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
- `enter` - View detailed container information (stats, config, env vars)
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

#### Container Details View
- `↑`/`↓` - Scroll through details
- `r` - Refresh stats
- `esc` - Return to container view

## Architecture

```
dockit/
├── main.go                # Application entry point
├── docker/                # Docker client wrapper
│   └── client.go
├── ui/                    # Bubble Tea UI components
│   ├── model.go           # Main application model
│   ├── dashboard.go       # Dashboard view
│   ├── containers.go      # Containers view
│   ├── container_details.go # Container details & stats view
│   ├── images.go          # Images view
│   ├── volumes.go         # Volumes view
│   ├── networks.go        # Networks view
│   ├── logs.go            # Logs viewer
│   ├── styles.go          # Lipgloss styles
│   └── keys.go            # Keybindings
└── components/            # Reusable components
```

## Technology Stack

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - Terminal UI framework using The Elm Architecture
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** - Style definitions and rendering
- **[Bubbles](https://github.com/charmbracelet/bubbles)** - Common TUI components
- **[Docker SDK](https://github.com/docker/docker)** - Official Docker Engine API client

## Screenshots

### Containers View
View and manage all your Docker containers with real-time status updates:
```
┌─ Containers ─┬─ Images ─┬─ Volumes ─┬─ Networks ─┐
│                                                    │
│  ● nginx-proxy      running    nginx:latest       │
│  ○ postgres-db      stopped    postgres:14        │
│  ● redis-cache      running    redis:alpine       │
└────────────────────────────────────────────────────┘
```

### Container Details View
Press `enter` on any container to view detailed stats and configuration:
```
STATISTICS
  CPU:         12.50%
  Memory:      245.3 MiB / 2048.0 MiB (11.98%)
  Network I/O: 1.2 MiB / 856.3 KiB
  Block I/O:   45.2 MiB / 12.1 MiB

ENVIRONMENT VARIABLES
  PATH=/usr/local/sbin:/usr/local/bin
  NODE_ENV=production
  DATABASE_URL=postgresql://...

CONFIGURATION
  Image:       nginx:latest
  Status:      Running
  Ports:       80/tcp, 443/tcp
  Port Bindings:
    80/tcp -> 0.0.0.0:8080
```

## Design Philosophy

Dockit is inspired by LazyDocker but with a fresh, modern aesthetic:
- Clean, card-based layouts
- Vibrant color scheme (purple, pink, cyan)
- Smooth vim-style navigation
- Intuitive keyboard shortcuts
- Real-time updates and live container stats

## Roadmap

- [ ] Docker Compose support
- [x] Volume management
- [x] Network management
- [x] Container stats (CPU, memory, I/O)
- [x] Container details inspection
- [x] Interactive help system
- [ ] Fuzzy search/filtering
- [ ] Bulk operations
- [ ] Export container configs
- [ ] Theme customization
- [ ] Configuration file support

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Acknowledgments

- Inspired by [LazyDocker](https://github.com/jesseduffield/lazydocker)
- Built with the amazing [Charm](https://charm.sh) ecosystem
