# Dockit - Project Overview

## What is Dockit?

Dockit is a transparent wrapper for Docker CLI that makes common commands prettier with colorful, well-formatted output. It's designed to be a drop-in replacement for `docker` that enhances readability without changing functionality.

## Core Concept

**Enhance what matters, pass through everything else.**

- ✅ Pretty formatting for common read commands (`ps`, `images`)
- ✅ Pass-through to Docker for everything else
- ✅ Zero configuration required
- ✅ 100% Docker CLI compatibility

## Current Status

**Phase 1: Foundation - COMPLETE** ✅

We've successfully implemented:
- Simple CLI wrapper architecture
- Pretty printer for `docker ps` with container IDs
- Pretty printer for `docker images`
- Clean format without side borders
- Consistent vertical dividers
- Colorful status indicators

## Architecture Overview

```
User Input
    ↓
main.go (router)
    ↓
    ├─→ "ps" → pretty.PrintContainers() → Formatted Output
    ├─→ "images" → pretty.PrintImages() → Formatted Output
    └─→ * → runDockerCommand() → Pass-through to Docker
```

### Key Components

**main.go**
- Entry point and command router
- Determines if command should be prettified or passed through
- Handles Docker command execution for pass-through

**pretty/containers.go**
- Fetches container list via Docker API
- Formats with colors and status indicators
- Shows: ID, name, status, image, ports, uptime

**pretty/images.go**
- Fetches image list via Docker API
- Formats with colors and human-readable sizes
- Shows: ID, repository:tag, size, creation time

## Design Philosophy

### 1. Transparent
Users should be able to replace `docker` with `dockit` in any command and get either:
- Enhanced pretty output (for supported commands)
- Standard Docker output (for unsupported commands)

### 2. Simple
- No complex TUI (terminal user interface)
- No interactive modes
- Just enhanced static output
- Fast execution

### 3. Clean
- No cluttered boxes or borders
- Only vertical dividers between columns
- Consistent formatting across all commands
- Pre-calculated spacing to avoid color code alignment issues

### 4. Colorful
- Green (●) for running/healthy
- Gray (○) for stopped/inactive
- Red (✖) for errors/failed
- Yellow (⏸) for paused/warning
- Cyan for headers
- Blue for primary identifiers

## Technical Implementation

### Color Handling
We use `github.com/fatih/color` for terminal colors. Key lesson learned: ANSI color codes interfere with padding, so we:
1. Pre-calculate all padding as strings
2. Print colored text without padding
3. Print padding separately

### Docker API
We use `github.com/docker/docker` official SDK:
- Direct API calls (faster than CLI parsing)
- Type-safe data structures
- Same data that Docker CLI uses

### Pass-through Strategy
For unhandled commands:
```go
cmd := exec.Command("docker", args...)
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
cmd.Stdin = os.Stdin
cmd.Run()
```

This ensures:
- All Docker features work
- Interactive commands work (like `exec -it`)
- Exit codes are preserved
- Streaming output works

## Output Format

### Standard Format
```
SECTION TITLE
──────────────────────────────────
[indicator] [id] │ [col1] │ [col2] │ [col3]
  ↪ Sub-info
  ⏱ Timestamp

Summary: X items (Y active)
```

### Example - Containers
```
CONTAINERS
──────────────────────────────────
● abc123def456 │ nginx-web │ running │ nginx:latest
  ↪ Ports: 8080:80/tcp
  ⏱ Up 2 hours
```

### Example - Images
```
IMAGES
──────────────────────────────────
abc123def456 │ nginx:latest │ 142.5 MB │ 2 days ago
```

## Dependencies

**Production:**
- `github.com/docker/docker` - Docker SDK for Go
- `github.com/fatih/color` - Terminal colors

**Development:**
- Go 1.24+
- Docker daemon

**Zero runtime dependencies** - single binary!

## Performance

Target: < 100ms overhead vs native Docker

Current performance:
- `dockit ps`: ~50ms overhead (API call + formatting)
- `dockit images`: ~50ms overhead (API call + formatting)
- Pass-through: ~10ms overhead (process spawn)

## Use Cases

### Primary Use Case
Daily Docker workflow where you frequently check containers and images:
```bash
dockit ps              # Quick visual scan of containers
dockit images          # Check image sizes
dockit logs myapp      # Pass-through for logs
dockit exec -it app bash   # Pass-through for exec
```

### Not Designed For
- Scripting/parsing (use `docker` directly or consider JSON output)
- CI/CD pipelines (use `docker` directly)
- Complex automation (use Docker API directly)

## What's Different from Other Tools?

### vs LazyDocker
- **LazyDocker**: Full TUI with interactive management
- **Dockit**: Simple command wrapper with pretty output
- **Key difference**: We're not replacing the workflow, just making output prettier

### vs Docker CLI
- **Docker CLI**: Functional but dense output
- **Dockit**: Same commands, prettier output
- **Key difference**: Drop-in replacement, zero learning curve

### vs Docker Desktop
- **Docker Desktop**: GUI application
- **Dockit**: Terminal-based CLI wrapper
- **Key difference**: For developers who live in the terminal

## Future Vision

**Phase 2: Enhanced Listings**
- Add `volume ls`, `network ls`, `system df`
- Maintain same clean format

**Phase 3: Interactive Enhancements**
- Colorized logs
- Pretty stats with progress bars
- Enhanced inspect output

**Phase 4: Advanced Features**
- Optional configuration file
- Custom color schemes
- Template system

**Long-term Vision:**
A complete prettier wrapper for all Docker read commands, while maintaining perfect pass-through for all write/action commands.

## Success Metrics

1. ✅ Drop-in replacement for `docker`
2. ✅ < 100ms overhead
3. ✅ Clean, consistent formatting
4. 🎯 Top 10 Docker commands prettified (2/10 complete)
5. 🎯 1000+ GitHub stars
6. 🎯 Featured in awesome-docker list

## Contributing

See [plan.md](plan.md) for implementation guidelines.

Quick guide:
1. Pick a command to prettify
2. Create `pretty/[command].go`
3. Add route in `main.go`
4. Follow existing format (no borders, vertical dividers)
5. Submit PR with examples

## Project History

- **2024-11-06**: Complete rewrite - simplified from complex TUI to simple wrapper
- **2024-11-05**: Initial version with complex Bubble Tea TUI
- **2024-11-06**: Current focus - clean, simple, fast wrapper

## License

MIT License - see LICENSE file for details
