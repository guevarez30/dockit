# Dockit Development Plan

## Project Overview

Dockit is a transparent wrapper for Docker CLI that makes common commands prettier with colorful, well-formatted output. For any command we haven't enhanced, we simply pass it through to the Docker CLI, ensuring 100% compatibility.

---

## Architecture

### Current Structure

```
dockit/
├── main.go              # Entry point and command router
├── pretty/              # Pretty printer implementations
│   ├── containers.go    # docker ps
│   └── images.go        # docker images
├── go.mod               # Dependencies
└── .claude/             # Documentation
    ├── roadmap.md       # Feature roadmap
    └── plan.md          # This file
```

### Design Principles

1. **Main.go is the router**
   - Parse first argument to determine command
   - Route to pretty printer if available
   - Pass through to Docker for everything else

2. **Pretty printers are self-contained**
   - Each command gets its own file in `pretty/`
   - Each printer handles its own Docker API calls
   - Consistent color scheme across all printers

3. **No breaking changes**
   - Must work as drop-in replacement for `docker`
   - Exit codes match Docker's
   - Stderr/stdout handled identically

---

## Implementation Guide

### Adding a New Pretty Command

Follow these steps to add a new pretty command:

#### 1. Create the Pretty Printer

Create a new file in `pretty/` (e.g., `pretty/volumes.go`):

```go
package pretty

import (
    "context"
    "fmt"
    "os"

    "github.com/docker/docker/client"
)

func PrintVolumes(args []string) {
    // 1. Create Docker client
    cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating Docker client: %v\n", err)
        os.Exit(1)
    }
    defer cli.Close()

    // 2. Call Docker API
    ctx := context.Background()
    volumes, err := cli.VolumeList(ctx, volume.ListOptions{})
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error listing volumes: %v\n", err)
        os.Exit(1)
    }

    // 3. Format and print with colors
    fmt.Println()
    cyan.Println("VOLUMES")
    cyan.Println(strings.Repeat("─", 90))

    for _, vol := range volumes.Volumes {
        // Pre-calculate padding to avoid color code alignment issues
        nameWidth := 40
        name := vol.Name
        if len(name) > nameWidth {
            name = name[:nameWidth-3] + "..."
        }
        namePadded := name + strings.Repeat(" ", nameWidth-len(name))

        // Print with colors and dividers
        blue.Print(namePadded)
        gray.Print(" │ ")
        fmt.Printf("%-20s", vol.Driver)
        gray.Print(" │ ")
        fmt.Println(vol.Scope)

        fmt.Println() // Blank line between items
    }

    // Summary
    fmt.Printf("Total: %d volumes\n", len(volumes.Volumes))
}
```

#### 2. Add Route in main.go

Add a case to the switch statement:

```go
switch command {
case "ps":
    pretty.PrintContainers(os.Args[2:])
case "images":
    pretty.PrintImages(os.Args[2:])
case "volume":
    // Check if it's 'volume ls'
    if len(os.Args) > 2 && os.Args[2] == "ls" {
        pretty.PrintVolumes(os.Args[3:])
    } else {
        runDockerCommand(os.Args[1:])
    }
default:
    runDockerCommand(os.Args[1:])
}
```

#### 3. Update Usage Text

Add the new command to `printUsage()` in main.go.

#### 4. Test

```bash
go build -o dockit
./dockit volume ls
```

---

## Color Scheme

Use these colors consistently across all pretty printers:

```go
var (
    green  = color.New(color.FgGreen, color.Bold)   // Success, running status
    red    = color.New(color.FgRed, color.Bold)     // Errors, stopped status
    yellow = color.New(color.FgYellow, color.Bold)  // Warnings, paused status
    cyan   = color.New(color.FgCyan, color.Bold)    // Borders, headers
    blue   = color.New(color.FgBlue, color.Bold)    // Primary info (names)
    gray   = color.New(color.FgHiBlack)             // Secondary info (timestamps, etc.)
)
```

### Color Guidelines

- **Green (●)**: Running containers, healthy status
- **Red (✖)**: Errors, failed status
- **Yellow (⏸)**: Paused, warning states
- **Cyan**: All borders and section headers
- **Blue**: Primary identifiers (container names, image names)
- **Gray**: Secondary information (timestamps, IDs)

---

## Output Format Guidelines

### Standard Format

All pretty printers should follow this format:

```
SECTION TITLE
──────────────────────────────────────────────────
● [id] │ [column1]  │ [column2] │ [column3]
  ↪ Sub-info: value
  ⏱ Timestamp info

○ [id] │ [column1]  │ [column2] │ [column3]
  ↪ Sub-info: value
  ⏱ Timestamp info

Summary: X items (Y active)
```

**Important Rules:**
- NO side borders (no │ on left/right edges)
- Only vertical dividers (│) between columns
- Title and horizontal line at top
- Pre-calculate padding to avoid color code alignment issues
- Blank line between items for readability

### Box Drawing Characters

Use these Unicode characters:
- `─` - Horizontal line (for header underline)
- `│` - Vertical line (for column dividers only, NOT borders)

### Status Indicators

Use these Unicode characters consistently:
- `●` - Active/Running (green)
- `○` - Inactive/Stopped (gray)
- `✖` - Error/Failed (red)
- `⏸` - Paused (yellow)
- `↪` - Sub-item indicator
- `⏱` - Time/timestamp indicator

---

## Testing Strategy

### Manual Testing

For each pretty printer:
1. Test with no items (empty list)
2. Test with one item
3. Test with many items (10+)
4. Test with very long names (truncation)
5. Test with all flags (`-a`, etc.)

### Pass-through Testing

Ensure pass-through works:
```bash
# These should behave exactly like docker
dockit run -d nginx
dockit build -t test .
dockit exec -it test bash
```

### Error Handling

Test error scenarios:
- Docker daemon not running
- Invalid flags
- Network errors
- Permission errors

---

## Next Commands to Implement

### Priority Order

1. **docker volume ls** (Phase 2)
   - Simple list format
   - Show driver and mountpoint
   - Similar to containers/images

2. **docker network ls** (Phase 2)
   - List networks with driver
   - Color-code by type
   - Show connected containers

3. **docker system df** (Phase 2)
   - Disk usage overview
   - Show reclaimable space
   - Color-coded warnings

4. **docker stats** (Phase 3)
   - Real-time stats
   - Progress bars for usage
   - Colorful output

5. **docker logs** (Phase 3)
   - Log level highlighting
   - Timestamp formatting
   - Maintain streaming capability

---

## Performance Considerations

### Benchmarks

Target performance overhead:
- `dockit ps`: < 50ms vs docker ps
- `dockit images`: < 50ms vs docker images
- Pass-through: < 10ms overhead

### Optimization Tips

1. **Minimize API calls** - One API call per command
2. **Efficient formatting** - Pre-allocate strings when possible
3. **Lazy loading** - Only fetch data that will be displayed
4. **Concurrent operations** - Use goroutines for independent operations

---

## Important Implementation Notes

### Color Code Alignment Issue

**Problem:** ANSI color codes interfere with string padding/alignment
- `Printf("%-30s", coloredString)` doesn't work correctly
- The padding counts color codes as characters

**Solution:** Pre-calculate padding, then print colored text
```go
// ❌ WRONG - alignment will be off
cyan.Printf("│ ")
blue.Printf("%-30s", name)
fmt.Printf(" │ ")

// ✅ CORRECT - pre-calculate padding
nameWidth := 30
namePadded := name + strings.Repeat(" ", nameWidth-len(name))
blue.Print(namePadded)
gray.Print(" │ ")
```

**Key principle:** Calculate padding based on visible text length, not the length of colored strings.

### Side Borders

**Don't use side borders:**
```go
// ❌ WRONG - creates side borders
cyan.Print("│ ")
// ... content ...
cyan.Println(" │")

// ✅ CORRECT - no side borders
// ... content with dividers between columns ...
fmt.Println()
```

### Vertical Dividers

**Use gray vertical dividers between columns:**
```go
gray.Print(" │ ")  // Between columns
```

---

## Dependencies

### Current Dependencies

```go
require (
    github.com/docker/docker v28.5.1+incompatible  // Docker API client
    github.com/fatih/color v1.18.0                 // Terminal colors
)
```

### Why These Dependencies?

- **docker/docker**: Official Docker SDK for Go - required for Docker API access
- **fatih/color**: Simple, widely-used color library for terminal output

### Adding New Dependencies

Before adding a new dependency:
1. Ensure it's necessary (can't implement ourselves)
2. Check license compatibility (MIT preferred)
3. Verify it's actively maintained
4. Consider binary size impact

---

## Code Style

### Go Best Practices

- Use `gofmt` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Keep functions under 50 lines
- Descriptive variable names
- Comment exported functions

### Error Handling

Always handle errors explicitly:

```go
if err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
```

Match Docker's exit codes:
- `0` - Success
- `1` - General error
- `125` - Docker daemon error
- `126` - Command cannot be invoked
- `127` - Command not found

---

## Release Process

### Version Bumping

Follow semantic versioning:
- `v0.x.y` - Pre-1.0 releases
- `v1.x.y` - Stable releases

Bump versions:
- **Patch** (0.1.1): Bug fixes
- **Minor** (0.2.0): New features
- **Major** (1.0.0): Breaking changes or v1.0 milestone

### Release Checklist

- [ ] All tests pass
- [ ] Update README with new features
- [ ] Update ROADMAP.md
- [ ] Tag release: `git tag v0.x.y`
- [ ] Build binaries for major platforms
- [ ] Write release notes

---

## Contributing Guidelines

### How to Contribute

1. **Pick an issue** from roadmap.md
2. **Create a branch**: `git checkout -b feature/volume-ls`
3. **Implement the feature** following this plan
4. **Test thoroughly** (manual + edge cases)
5. **Submit PR** with screenshots of output

### PR Requirements

- Clear description of what's changed
- Screenshots of new pretty output
- Confirmation that pass-through still works
- Updated README if needed

---

## Future Architecture Considerations

### Plugin System (Future)

For v2.0, consider:
- Plugin architecture for custom printers
- User-defined output templates
- Theme system

### Configuration (Future)

Potential `~/.dockitrc`:
```json
{
  "colors": {
    "running": "green",
    "stopped": "red"
  },
  "format": {
    "dateFormat": "relative",
    "truncateNames": true
  }
}
```

---

## Maintenance

### Regular Tasks

- Update Docker SDK when new versions release
- Monitor for security vulnerabilities
- Review and merge community PRs
- Keep documentation in sync with code

### Support

- GitHub Issues for bug reports
- Discussions for feature requests
- Examples in README for common use cases

---

## Success Criteria

A successful implementation:
1. ✅ Looks better than standard Docker output
2. ✅ Works as drop-in replacement for Docker
3. ✅ Adds minimal overhead (< 100ms)
4. ✅ Handles errors gracefully
5. ✅ Maintains pass-through for unhandled commands
