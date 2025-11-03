# Task Tool Specification

## Mission

Build a minimalist, local-first task management tool optimized for LLM-driven development workflows. The tool should be so fast and simple that it becomes invisible—replacing ad-hoc TODO comments and markdown files with structured, persistent project memory.

## Design Principles

### The Torvalds Principle
At every juncture, optimize for speed and simplicity. This tool should be so streamlined and efficient that alternatives feel unnecessary. Target: <10ms response time for all commands.

### Core Values
1. **Speed First** - Sub-10ms response time, no perceptible latency
2. **Simplicity** - Fewer features, better execution
3. **LLM-Native** - Designed for how LLMs think and work
4. **Repo-Local** - Lives with the code, version controlled
5. **Zero Dependencies** - Single binary, works everywhere
6. **Git-Friendly** - Merge conflicts should be rare and obvious
7. **Offline-First** - No network, no services, no accounts

## The Problem We're Solving

### Current State (Broken)
- LLMs create scattered TODO.md, NOTES.md, PLAN.md files
- Each conversation starts fresh after context compaction
- Multiple LLMs working on same repo create chaos
- No shared source of truth
- Project knowledge degrades over time
- Decision archaeology is impossible

### Desired State (Fixed)
- Single structured task system per repo
- Persistent project memory across sessions
- All LLMs (Claude, GPT, etc.) read/write same data
- Humans can see LLM's "internal TODO" and comment on it
- Project gets BETTER over time as task history grows
- Clear timeline of decisions and changes

## User Experience

### Installation (One-Time, Global)
```bash
# Single command installation
curl -sSL https://tasks.onuse.dev/install | sh

# Downloads correct binary for OS/arch to ~/.local/bin/task
# No dependencies, no runtime, just works
```

### Per-Repository Setup
```bash
cd my-project
task init
# Creates .tasks/ directory structure
# User commits to git
```

### Daily Usage (Human)
```bash
task list                          # See all active tasks
task create "Refactor auth" "Move to OAuth2"
task update 42 --status done
task show 42                       # Full task details
task list --status done            # Filter by status
```

### Daily Usage (LLM)
LLMs use identical commands. In project instructions:

```markdown
Use the `task` tool to track work instead of markdown TODO lists.

After context compaction, run: task context
Before starting work, run: task list --status active
When creating new work, run: task create "title" "description"
When completing work, run: task update ID --status done
Add notes with: task update ID --note "your note here"
```

The `task context` command provides a compact project overview optimized for LLM consumption after context resets.

## Technical Specification

### Language & Distribution
- **Language:** Go (for speed, single binary, cross-platform)
- **Distribution:** GitHub Releases with binaries for:
  - darwin-amd64, darwin-arm64
  - linux-amd64, linux-arm64
  - windows-amd64
- **Target:** Single binary <5MB, zero runtime dependencies

### File Structure

```
.tasks/
  manifest.json          # Just next_id counter
  tasks/
    00001.json          # One file per task
    00002.json
    00003.json
  index.json           # Cached index for fast queries (mandatory)
```

**Why this structure:**
- New tasks = new files = no merge conflicts
- Updates to different tasks = different files = no conflicts
- Only manifest.json needs locking (trivial conflict resolution)
- Git diffs are meaningful
- Grep-able and human-readable
- index.json rebuilt on every write operation to guarantee <10ms performance

### Data Model

#### Task Structure
```json
{
  "id": 1,
  "created": "2025-11-03T10:30:00Z",
  "updated": "2025-11-03T14:20:00Z",
  "status": "active",
  "title": "Refactor authentication system",
  "description": "Move from basic auth to OAuth2. Need to support Google and GitHub providers.",
  "notes": [
    {
      "timestamp": "2025-11-03T14:20:00Z",
      "author": "claude",
      "text": "Started implementation, created oauth package"
    }
  ],
  "dependencies": [12, 15],
  "tags": ["security", "refactor"]
}
```

#### Manifest Structure
```json
{
  "next_id": 43,
  "created": "2025-11-03T10:00:00Z",
  "version": "1.0"
}
```

#### Index Structure
```json
{
  "tasks": [
    {
      "id": 1,
      "status": "active",
      "title": "Refactor authentication system",
      "created": "2025-11-03T10:30:00Z",
      "updated": "2025-11-03T14:20:00Z"
    },
    {
      "id": 2,
      "status": "backlog",
      "title": "Add rate limiting",
      "created": "2025-11-03T11:00:00Z",
      "updated": "2025-11-03T11:00:00Z"
    }
  ],
  "updated": "2025-11-03T14:20:00Z"
}
```

**Note:** This index is regenerated on every create/update operation to ensure fast queries.

#### Status Values
- `backlog` - Not started
- `active` - Currently being worked on
- `done` - Completed
- `cancelled` - Won't do

### Command Line Interface

#### Core Commands (MVP)

```bash
task init
# Creates .tasks/ directory structure in current repo
# Exit with error if already exists

task create <title> [description]
# Creates new task in backlog status
# Returns: "Created task #42"
# Exit code 0 on success

task list [--status STATUS] [--format FORMAT]
# Lists tasks, defaults to active only
# --status: backlog, active, done, cancelled, all
# --format: text (default), json, compact
# Output sorted by ID ascending

task show <id>
# Shows full task details including all notes
# Exit code 1 if task not found

task update <id> [--status STATUS] [--note NOTE] [--title TITLE] [--description DESC]
# Updates task fields
# --status: changes status
# --note: appends timestamped note
# Multiple flags can be combined
# Exit code 1 if task not found

task context [--format FORMAT]
# Outputs project context optimized for LLM consumption
# --format: text (default), json
# Shows active tasks, recent completions, blocked tasks
# Designed to be included in LLM system prompts after context compaction
```

#### Future Commands (Post-MVP)
```bash
task serve [--port PORT]
# Starts local web UI server (embedded HTML)

task search <query>
# Full-text search across tasks

task depend <id> <dep_id>
# Add dependency relationship

task export [--format FORMAT]
# Export to JSON, CSV, markdown
```

### Command Output Design

**Principle:** Output should be parseable by both humans and LLMs.

**List format (text):**
```
#42  [active]   Refactor authentication system
#43  [backlog]  Add rate limiting
#44  [active]   Update documentation
```

**List format (json):**
```json
[
  {
    "id": 42,
    "status": "active",
    "title": "Refactor authentication system",
    "created": "2025-11-03T10:30:00Z"
  }
]
```

**Show format:**
```
Task #42: Refactor authentication system
Status: active
Created: 2025-11-03 10:30:00
Updated: 2025-11-03 14:20:00

Description:
Move from basic auth to OAuth2. Need to support Google and GitHub
providers.

Dependencies: #12, #15

Notes:
  [2025-11-03 14:20] claude: Started implementation, created oauth package
  [2025-11-03 15:10] human: Don't forget to add tests
```

**Context format (text):**
```
PROJECT CONTEXT

Active Tasks (3):
  #42  Refactor authentication system
  #43  Add rate limiting
  #44  Update documentation

Recently Completed (2):
  #40  Fix login bug (completed 2025-11-02)
  #41  Add logging (completed 2025-11-02)

Total: 15 tasks (3 active, 8 backlog, 4 done, 0 cancelled)
```

**Context format (json):**
```json
{
  "active": [
    {"id": 42, "title": "Refactor authentication system"},
    {"id": 43, "title": "Add rate limiting"},
    {"id": 44, "title": "Update documentation"}
  ],
  "recently_completed": [
    {"id": 40, "title": "Fix login bug", "completed": "2025-11-02T16:30:00Z"},
    {"id": 41, "title": "Add logging", "completed": "2025-11-02T17:15:00Z"}
  ],
  "summary": {
    "total": 15,
    "active": 3,
    "backlog": 8,
    "done": 4,
    "cancelled": 0
  }
}
```

### Error Handling

**Exit Codes:**
- 0: Success
- 1: Generic error (task not found, invalid input)
- 2: File system error (can't read/write)
- 3: Not in a task-enabled repo (no .tasks/ directory)

**Error Messages:**
```
Error: not in a task repository (no .tasks/ directory found)
Run 'task init' to initialize task tracking in this repository.

Error: task #99 not found

Error: invalid status "foo" (must be: backlog, active, done, cancelled)
```

### Performance Targets

| Operation | Target | Maximum |
|-----------|--------|---------|
| task list | <5ms | 10ms |
| task create | <5ms | 10ms |
| task show | <3ms | 10ms |
| task update | <5ms | 10ms |
| task context | <5ms | 10ms |
| Binary size | <5MB | 10MB |
| Startup time | <1ms | 5ms |

**Why this matters:**
- LLMs may call `task list` 100+ times per session
- 100 calls × 50ms = 5 seconds of pure overhead
- 100 calls × 5ms = 0.5 seconds (imperceptible)

### Web UI (Post-MVP)

A single-file HTML interface embedded in the Go binary:

```bash
task serve --port 8080
# Starts HTTP server
# Navigate to http://localhost:8080
# Shows interactive task board
```

**Features:**
- View all tasks in a Kanban-style board
- Filter by status, search
- Click task to see details
- Read-only initially (editing via CLI only)
- No build step, no dependencies, just vanilla HTML/CSS/JS
- Auto-refreshes when .tasks/ files change

## Project Structure

```
tasks/
  main.go                  # CLI entry point
  go.mod
  go.sum
  
  internal/
    store/
      store.go            # File I/O operations
      manifest.go         # Manifest handling
    task/
      task.go             # Task struct and methods
      validation.go       # Input validation
    commands/
      init.go             # task init
      create.go           # task create
      list.go             # task list
      show.go             # task show
      update.go           # task update
      context.go          # task context
    ui/
      web.go              # Embedded web UI (post-MVP)
  
  docs/
    SPEC.md               # This file
    CLI.md                # Command reference
  
  examples/
    .clinerules           # Example LLM instructions
    custom-filter.sh      # Script examples
  
  .github/
    workflows/
      release.yml         # Build multi-platform binaries
  
  README.md
  LICENSE
```

## Development Phases

### Phase 1: Core MVP (4 hours)
- [ ] Project setup, go.mod
- [ ] Data structures (Task, Manifest, Index)
- [ ] File storage layer (read/write JSON, index rebuilding)
- [ ] Commands: init, create, list, show, update, context
- [ ] Basic error handling
- [ ] Manual testing

### Phase 2: Polish (2 hours)
- [ ] Better error messages
- [ ] Input validation
- [ ] --format flags
- [ ] Unit tests for core functions
- [ ] README with examples

### Phase 3: Distribution (2 hours)
- [ ] GitHub Actions for releases
- [ ] Multi-platform binary builds
- [ ] Install script (tasks.onuse.dev/install)
- [ ] Documentation site

### Phase 4: Post-MVP Features
- [ ] Web UI (task serve)
- [ ] Full-text search
- [ ] Dependency tracking
- [ ] Export formats
- [ ] Git hooks integration

## Testing Strategy

### Manual Testing Checklist
```bash
# Setup
cd /tmp/test-repo && git init
task init
git add .tasks && git commit -m "init tasks"

# Create tasks
task create "First task" "Description here"
task create "Second task"
task list

# Update tasks
task update 1 --status active
task update 1 --note "Started work"
task show 1
task update 1 --status done
task list --status done

# Edge cases
task show 999  # Should error gracefully
task update 1 --status invalid  # Should error with helpful message
task create ""  # Should require title
```

### Performance Testing
```bash
# Create 1000 tasks
for i in {1..1000}; do
  task create "Task $i" "Description $i"
done

# Benchmark list
time task list  # Should be <10ms

# Benchmark show
time task show 500  # Should be <5ms
```

## Integration with LLMs

### Example .clinerules / Project Instructions

```markdown
## Task Management

This project uses the `task` tool for structured work tracking.

### After Context Compaction
Get project overview:
```bash
task context
```

### Before Starting Work
Always check current tasks:
```bash
task list --status active
```

### Creating Tasks
When identifying new work, create a task:
```bash
task create "Brief title" "Detailed description"
```

### Updating Tasks
When working on a task:
```bash
task update <id> --note "Progress update"
```

When completing:
```bash
task update <id> --status done
```

### Principles
- Use tasks instead of TODO comments in code
- Use tasks instead of creating TODO.md files
- Break large work into multiple tasks
- Keep task titles short and actionable
- Add notes as you make progress
- Run `task context` at session start to restore project state
```

## Success Metrics

### Adoption Indicators
- Developers install once, use daily
- LLMs naturally use task commands without prompting
- Tasks persist and remain useful across weeks/months
- Zero "how do I use this?" questions after reading README
- Tool "disappears" - users don't think about it

### Technical Indicators
- All commands respond in <10ms
- Binary size <5MB
- Zero reported bugs related to file corruption
- Works identically on Mac/Linux/Windows
- No dependencies beyond stdlib

### Ecosystem Indicators
- Other tools built on top (custom UIs, integrations)
- Users write shell aliases and helper scripts
- Appears in .clinerules / Claude project templates
- Referenced in LLM best practices guides

## Non-Goals

**This tool explicitly does NOT:**
- ❌ Sync across machines (use git)
- ❌ Support teams/permissions (it's local files)
- ❌ Integrate with Jira/Linear/etc (it's independent)
- ❌ Have a mobile app (it's CLI-first)
- ❌ Track time automatically (add notes manually)
- ❌ Support plugins (fork and modify instead)
- ❌ Have a cloud service (local forever)
- ❌ Require configuration files (works out of box)

**Philosophy:** Do one thing well. Be the Git of task tracking for LLM workflows.

## Implementation Notes

### Index Management
- index.json is regenerated on every create/update operation
- Write new index atomically (temp file + rename)
- Index enables fast filtering and summary generation
- If index is missing/corrupted, rebuild from task files
- Keep index minimal (no descriptions, notes, dependencies)

### Concurrency Considerations
- Use file locking when modifying manifest.json (next_id counter)
- Task files are append-mostly (rare conflicts)
- Index rebuilt atomically on each write
- No need for database or complex locking

### Cross-Platform Compatibility
- Use filepath.Join() for paths
- Test on Windows (different path separators)
- Handle line endings (CRLF vs LF)

### Error Recovery
- If manifest.json corrupted, can rebuild from task files
- Validate JSON on read, fail fast
- Never partially write files (write to temp, then rename)

### Future Extensibility
- Keep data format simple (others can build tools)
- JSON is universal (any language can parse)
- File-per-task allows parallel tools
- Web UI is optional enhancement

## FAQ for Implementers

**Q: Why not use SQLite?**
A: Files are simpler, grep-able, git-friendly, and fast enough.

**Q: Why JSON not YAML?**
A: JSON is stdlib, YAML requires dependencies. Simplicity wins.

**Q: Why not support task priorities?**
A: Keep it simple. Users can add as needed via notes or custom fields.

**Q: Should we support sub-tasks?**
A: Post-MVP. Dependencies cover most cases initially.

**Q: What about concurrent edits?**
A: Rare in practice. Different tasks = different files = no conflict. Manifest has simple last-write-wins.

**Q: Should the web UI allow editing?**
A: Eventually, but read-only is fine for MVP. CLI is primary interface.

---

## Getting Started (For Implementation)

1. **Set up Go project:** `go mod init github.com/onuse/tasks`
2. **Start with data structures:** Define Task and Manifest structs
3. **Build storage layer:** Read/write JSON files
4. **Implement commands one by one:** init → create → list → show → update
5. **Test manually at each step:** Don't build everything before testing
6. **Optimize after it works:** Make it correct, then make it fast

## Reference Implementation Checklist

- [ ] Go 1.21+ (for latest stdlib features)
- [ ] Use `encoding/json` for serialization
- [ ] Use `flag` or `cobra` for CLI parsing
- [ ] Use `os` and `path/filepath` for file operations
- [ ] Use `time` package for timestamps (RFC3339 format)
- [ ] Keep dependencies minimal (prefer stdlib)
- [ ] Format code with `gofmt`
- [ ] Include `-ldflags` for version info in binary

---

**This specification is complete and sufficient to build the tool. Any ambiguity should be resolved by preferring simplicity and speed.**
