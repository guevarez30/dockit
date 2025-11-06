# Dockit 🐳

A simple, prettier wrapper for Docker CLI commands. Dockit enhances common Docker commands with beautiful, colorful terminal output while maintaining full compatibility with the standard Docker CLI.

## Philosophy

Dockit is a transparent wrapper around Docker. It makes common commands prettier, but for everything else, it simply passes through to the standard Docker CLI. If we haven't built out pretty formatting for a command yet, you'll get the traditional Docker response.

## Features

- **Pretty `docker ps`** - Beautiful, colorful container listings with status indicators
- **Pretty `docker images`** - Enhanced image listings with formatted sizes and timestamps
- **Full Docker Compatibility** - All other Docker commands work exactly as they do with `docker`
- **Zero Configuration** - Works out of the box with your existing Docker setup

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

Use `dockit` exactly like you would use `docker`:

```bash
# Pretty commands (enhanced output)
dockit ps                    # List running containers with colors
dockit ps -a                 # List all containers with colors
dockit images                # List images with pretty formatting

# All other commands pass through to docker
dockit run -d nginx          # Standard docker run
dockit build -t myapp .      # Standard docker build
dockit exec -it web bash     # Standard docker exec
dockit logs myapp            # Standard docker logs
```

### Command Reference

**Pretty Commands** (enhanced with colors and formatting):
- `dockit ps [-a]` - List containers
- `dockit images` - List images

**Pass-through Commands** (standard Docker output):
- All other Docker commands work as normal: `run`, `build`, `exec`, `logs`, `pull`, `push`, `stop`, `start`, `rm`, `rmi`, etc.

## Examples

### Pretty Container Listing
```bash
$ dockit ps

╭─────────────────────────────────────────────────────────────────────────────────╮
│ CONTAINERS                                                                      │
├─────────────────────────────────────────────────────────────────────────────────┤
│ ● nginx-web              │ running    │ nginx:latest                           │
│   ↪ Ports: 80:8080/tcp                                                         │
│   ⏱ Up 2 hours                                                                  │
├─────────────────────────────────────────────────────────────────────────────────┤
│ ○ postgres-db            │ exited     │ postgres:14                            │
│   ⏱ Exited (0) 5 minutes ago                                                   │
├─────────────────────────────────────────────────────────────────────────────────┤
╰─────────────────────────────────────────────────────────────────────────────────╯

Total: 2 containers (1 running)
```

### Pretty Image Listing
```bash
$ dockit images

╭─────────────────────────────────────────────────────────────────────────────────╮
│ IMAGES                                                                          │
├─────────────────────────────────────────────────────────────────────────────────┤
│ nginx:latest                          │ 142.5 MB   │ abc123def456            │
│   ⏱ Created: 2 days ago                                                        │
├─────────────────────────────────────────────────────────────────────────────────┤
│ postgres:14                           │ 376.2 MB   │ def456ghi789            │
│   ⏱ Created: 1 week ago                                                        │
├─────────────────────────────────────────────────────────────────────────────────┤
╰─────────────────────────────────────────────────────────────────────────────────╯

Total: 2 images (Total size: 518.7 MB)
```

### Pass-through Commands
```bash
# These work exactly like docker
dockit run -d -p 8080:80 nginx
dockit build -t myapp:latest .
dockit exec -it mycontainer bash
dockit logs -f myapp
```

## Architecture

```
dockit/
├── main.go              # CLI entry point and command router
├── pretty/              # Pretty printers for enhanced commands
│   ├── containers.go    # Pretty printer for 'docker ps'
│   └── images.go        # Pretty printer for 'docker images'
└── .claude/             # Development documentation
    ├── roadmap.md       # Feature roadmap
    └── plan.md          # Development plan
```

## Roadmap

See [.claude/roadmap.md](.claude/roadmap.md) for the complete feature roadmap.

### Upcoming Pretty Commands
- `dockit volumes` - Pretty volume listing
- `dockit networks` - Pretty network listing
- `dockit stats` - Enhanced real-time stats
- `dockit logs` - Colorized log output

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Why Dockit?

Docker's CLI output is functional but dense. Dockit makes it easier to scan and understand your Docker environment at a glance while maintaining 100% compatibility with the Docker CLI you know and love.
