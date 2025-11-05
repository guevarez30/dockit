# Dockit - Feature Reference

## Complete Feature List

### Dashboard View
**Access**: Tab key or launch application (default view)

**Features**:
- Container statistics card
  - Total containers count
  - Running containers (green)
  - Stopped containers (red)
- Image statistics card
  - Total images count
  - Dangling images count

**Keybindings**:
- `tab` - Switch to Containers view
- `q` - Quit application

---

### Containers View
**Access**: Tab key from Dashboard or Images

**Display Columns**:
1. **STATUS** - Visual indicator (● color + state text)
   - Green ● running
   - Red ● exited
   - Yellow ● paused
2. **NAME** - Container name (truncated at 25 chars)
3. **IMAGE** - Image name (truncated at 30 chars)
4. **ID** - Short container ID (12 chars)
5. **UPTIME** - Time since started/created
   - "30s ago", "5m ago", "2h ago", "3d ago"
   - "stopped" for exited containers

**Operations**:
- `s` - Start selected container
  - Shows "⟳ Processing..." during action
  - Shows "✓ Container started" on success
  - Auto-refreshes list
- `x` - Stop selected container
  - Shows "⟳ Processing..."
  - Shows "✓ Container stopped" on success
- `r` - Restart selected container
  - Shows "⟳ Processing..."
  - Shows "✓ Container restarted" on success
  - Uptime resets after restart
- `d` - Remove selected container (force)
  - Shows "⟳ Processing..."
  - Shows "✓ Container removed" on success
- `L` - View logs (capital L)
  - Opens full-screen log viewer
- `ctrl+r` - Refresh container list manually

**Navigation**:
- `↑/k` - Move selection up
- `↓/j` - Move selection down
- `tab` - Switch to Images view
- `q` - Quit application

**Visual Feedback**:
- Selected row: Purple background
- Status messages: Green with ✓ icon
- Processing indicator: Yellow with ⟳ icon
- Messages auto-clear after 2 seconds

---

### Images View
**Access**: Tab key from Containers

**Display Columns**:
1. **REPOSITORY:TAG** - Full image tag (truncated at 40 chars)
   - Shows "<none>:<none>" for untagged images
2. **IMAGE ID** - Short image ID (12 chars, sha256: prefix removed)
3. **SIZE** - Image size in MB (e.g., "125.3 MB")
4. **[dangling]** indicator - Yellow text for dangling images

**Operations**:
- `d` - Remove selected image (force)
  - Shows "⟳ Processing..."
  - Shows "✓ Image removed" on success
  - Auto-refreshes list
- `ctrl+r` - Refresh image list manually

**Navigation**:
- `↑/k` - Move selection up
- `↓/j` - Move selection down
- `tab` - Switch to Compose view (placeholder) → Dashboard
- `q` - Quit application

---

### Logs View
**Access**: Press `L` on any container in Containers view

**Display**:
- Full-screen log viewer
- Title shows container ID (12 chars)
- Scrollable viewport
- Last 500 lines loaded

**Search Mode**:
Activated by pressing `/`

**Search Features**:
1. Enter search term in input box
2. Press `enter` to apply search
3. Results show:
   - Only matching lines displayed
   - Search term highlighted (orange/yellow background)
   - Match count displayed: "Filtered: X matches for 'term'"
4. Case-insensitive matching
5. Multiple occurrences per line highlighted

**Keybindings**:

*Normal Mode* (viewing logs):
- `↑/k` - Scroll up
- `↓/j` - Scroll down
- `/` - Enter search mode
- `esc` - Clear filter (if filtered) OR exit logs (if not filtered)
- `q` - Does NOT work in logs (use esc)

*Search Mode* (entering search):
- Type characters - Add to search term
- `enter` - Apply search filter
- `esc` - Cancel search (stay in logs, no filter)

*Filtered Mode* (after search):
- `↑/k` - Scroll filtered results
- `↓/j` - Scroll filtered results
- `/` - Start new search
- `esc` - Clear filter, show all logs
- `esc` (again) - Exit logs view

**Search Behavior**:
- Empty search shows all logs
- "No matches found" if term not in logs
- Highlights preserved while scrolling
- Search state cleared when exiting logs

**Log Format**:
- Parses Docker's 8-byte header format
- Falls back to line-by-line if parsing fails
- Strips timestamps (Docker adds them)
- Shows "No logs available" if empty

---

## Keyboard Reference

### Global (All Views)
| Key | Action |
|-----|--------|
| `tab` | Switch view (Dashboard → Containers → Images → Dashboard) |
| `q` | Quit application (except in Logs) |
| `?` | Show help (planned) |

### Navigation
| Key | Action |
|-----|--------|
| `↑` or `k` | Move up |
| `↓` or `j` | Move down |
| `←` or `h` | Move left (reserved) |
| `→` or `l` | Move right (reserved) |

### Container Operations
| Key | Action |
|-----|--------|
| `s` | Start container |
| `x` | Stop container |
| `r` | Restart container |
| `d` | Delete container/image |
| `L` | View logs (capital L) |
| `ctrl+r` | Refresh list |

### Log Viewer
| Key | Action |
|-----|--------|
| `/` | Search logs |
| `esc` | Clear filter / Exit logs |
| `↑/↓` | Scroll |

---

## Feature Evolution & Iterations

### Issues Fixed During Development

1. **Tabs Not Showing**
   - Problem: Tabs weren't visible on Containers/Images views
   - Cause: Title sections were removed, tabs got lost
   - Solution: Ensured tabs render in main model before content
   - Added separator line for clarity

2. **Tab Size Inconsistency**
   - Problem: Active tab much larger than inactive
   - Cause: Active had rounded border, inactive didn't
   - Solution: Removed border, kept consistent padding

3. **Container Display Messy**
   - Problem: Long names/images wrapped, unreadable
   - Cause: No fixed column widths
   - Solution: Truncation + fixed-width columns

4. **Logs Stuck on "Loading..."**
   - Problem: Logs never displayed
   - Cause: `m.ready` only set on window resize
   - Solution: Set `m.ready = true` when logs received

5. **Restart No Feedback**
   - Problem: User couldn't tell if restart worked
   - Cause: No visual feedback system
   - Solution: Added status messages + progress indicators

6. **Docker Log Format Issues**
   - Problem: Logs showed garbled text or empty
   - Cause: Docker's 8-byte header not parsed
   - Solution: Custom parser with fallback

---

## Planned Features (Not Yet Implemented)

### Docker Compose Support
- View compose projects
- Up/down services
- Service logs
- Scale services

### Volume Management
- List volumes
- Remove volumes
- Volume usage stats

### Network Inspection
- List networks
- Network details
- Connected containers

### Container Stats
- Real-time CPU usage
- Memory usage graphs
- Network I/O
- Disk I/O

### Advanced Features
- Fuzzy search across views
- Bulk operations (multi-select)
- Theme customization
- Config file (~/.dockitrc)
- Export logs to file
- Container shell access

---

## Known Limitations

1. **No Compose Support** - Placeholder tab exists
2. **Sync Operations** - UI blocks during Docker API calls
3. **No Auto-Refresh** - Must manually refresh (ctrl+r)
4. **Limited Log Lines** - Shows last 500 lines only
5. **No Log Streaming** - Static snapshot, not live
6. **Force Delete Only** - No confirmation dialogs
7. **No Multi-Select** - Can't act on multiple containers
8. **No Config File** - Hardcoded settings
