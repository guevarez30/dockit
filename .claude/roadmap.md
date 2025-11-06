# Dockit Roadmap

## Vision
Dockit is a transparent, prettier wrapper for Docker CLI. We enhance common Docker commands with beautiful, colorful output while maintaining 100% compatibility with the Docker CLI.

## Core Principles
1. **Transparent** - Pass through to Docker for anything we don't enhance
2. **Simple** - No complex TUI, just prettier command output
3. **Compatible** - Works exactly like Docker for all unenhanced commands
4. **Beautiful** - Colorful, well-formatted output that's easy to scan
5. **Zero Config** - Works out of the box

---

## Phase 1: Foundation ✅ (Current)
**Status: Complete**

- [x] Simple CLI wrapper architecture
- [x] Pass-through to Docker for unhandled commands
- [x] Pretty printer for `docker ps`
- [x] Pretty printer for `docker images`
- [x] Color-coded output with status indicators
- [x] Updated documentation

### Deliverables
- Working `dockit ps` with colorful container listings
- Working `dockit images` with formatted image listings
- All other Docker commands pass through transparently

---

## Phase 2: Enhanced Listings (Next)
**Goal: Make more listing commands prettier**

### Planned Features
- [ ] Pretty printer for `docker volume ls`
  - Show volume name, driver, and scope
  - Format mountpoint locations
  - Color-code by driver type

- [ ] Pretty printer for `docker network ls`
  - Show network name, driver, and scope
  - Color-code by network type (bridge, host, etc.)
  - Display connected containers

- [ ] Pretty printer for `docker system df`
  - Enhanced disk usage visualization
  - Color-coded size warnings
  - Beautiful formatting for space statistics

### Success Criteria
- All basic listing commands have pretty output
- Output is easier to read than standard Docker
- Performance remains fast (< 100ms overhead)

---

## Phase 3: Interactive Enhancements
**Goal: Add light interactivity to static commands**

### Completed Features
- [x] Interactive `docker logs` TUI
  - Built-in search with regex support
  - Scroll through history (arrows, PgUp/PgDn, vim keys)
  - Follow mode for live streaming
  - Pause/resume capability
  - Jump to next/previous match
  - Preserves original log colors from containers

### Planned Features
- [ ] Pretty printer for `docker stats`
  - Real-time colorful stats output
  - Progress bars for CPU/memory usage
  - Formatted data rates

- [ ] Pretty printer for `docker inspect`
  - Formatted JSON with syntax highlighting
  - Collapsible sections
  - Human-readable timestamps

### Success Criteria
- Real-time commands maintain low latency
- Colors enhance readability
- Still works as standard pass-through if needed

---

## Phase 4: Advanced Features
**Goal: Add useful helpers without compromising simplicity**

### Planned Features
- [ ] `dockit status` - Dashboard view
  - Quick overview of containers, images, volumes
  - Disk usage summary
  - Running vs stopped containers

- [ ] Column width auto-detection
  - Adapt output to terminal width
  - Smart truncation for narrow terminals

- [ ] Configuration file support
  - Optional `~/.dockitrc` for customization
  - Color scheme preferences
  - Default flags for commands

- [ ] Table sorting options
  - Sort containers by name, status, image
  - Sort images by size, age, name

### Success Criteria
- Features are opt-in, not required
- No breaking changes to existing functionality
- Still feels simple and fast

---

## Future Considerations

### Potential Features (Backlog)
- JSON output option (for scripting)
- Custom output templates
- Filtering helpers (e.g., show only running containers)
- Integration with Docker Compose
- Multi-host support (remote Docker daemons)

### Non-Goals
- **No complex TUI** - We're not building lazydocker
- **No container management UI** - Use Docker CLI for actions
- **No custom Docker commands** - Just wrap existing ones
- **No configuration required** - Should work perfectly with zero setup

---

## Versioning Strategy

### v0.1.0 ✅
- Basic wrapper with `ps` and `images` pretty printers

### v0.2.0 (Current)
- Add interactive `logs` TUI with search and follow mode
- Next: `volumes`, `networks`, and `system df` pretty printers

### v0.3.0
- Add `stats` and `inspect` enhancements

### v1.0.0
- All core listing commands have pretty printers
- Stable API
- Full test coverage
- Performance benchmarks

---

## Success Metrics

1. **Performance**: < 100ms overhead vs native Docker commands
2. **Coverage**: Pretty printers for top 10 most-used Docker commands
3. **Adoption**: Users prefer dockit for daily Docker work
4. **Simplicity**: Codebase remains under 2000 LOC
5. **Compatibility**: 100% pass-through compatibility with Docker CLI

---

## Contributing

Want to help build Dockit? Here's how:

1. **Pick a command** from Phase 2 or 3
2. **Create a pretty printer** in `pretty/`
3. **Add the route** in `main.go`
4. **Submit a PR** with examples

See [plan.md](plan.md) for implementation details.
