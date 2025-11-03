# CLI Reference

Complete command-line interface documentation for the task management tool.

## Table of Contents

- [Global Options](#global-options)
- [Commands](#commands)
  - [init](#init)
  - [create](#create)
  - [list](#list)
  - [show](#show)
  - [update](#update)
  - [link](#link)
  - [unlink](#unlink)
  - [search](#search)
  - [context](#context)
  - [serve](#serve)

## Global Options

All commands return:
- Exit code `0` on success
- Exit code `1` on generic error (task not found, invalid input)
- Exit code `2` on file system error
- Exit code `3` when not in a task-enabled repository

## Commands

### init

Initialize task tracking in the current repository.

**Usage:**
```bash
task init
```

**Description:**
Creates the `.tasks/` directory structure in the current directory. This includes:
- `manifest.json` - Tracks the next task ID
- `index.json` - Cached index for fast queries
- `tasks/` directory - Individual task files

**Examples:**
```bash
cd my-project
task init
git add .tasks
git commit -m "Initialize task tracking"
```

**Errors:**
- Returns error if `.tasks/` already exists

---

### create

Create a new task.

**Usage:**
```bash
task create <title> [description]
```

**Arguments:**
- `title` (required) - Short title for the task
- `description` (optional) - Detailed description

**Description:**
Creates a new task in `backlog` status. Tasks are assigned sequential IDs starting from 1.

**Examples:**
```bash
# Simple task
task create "Fix login bug"

# With description
task create "Implement OAuth" "Add Google and GitHub authentication"

# Multi-word titles (use quotes)
task create "Update user documentation"
```

**Output:**
```
Created task #42
```

---

### list

List tasks with filtering and sorting options.

**Usage:**
```bash
task list [--status STATUS] [--sort FIELD] [--reverse] [--format FORMAT]
```

**Options:**
- `--status` - Filter by status (default: `active`)
  - Values: `backlog`, `active`, `blocked`, `done`, `cancelled`, `all`
- `--sort` - Sort tasks by field (default: `id`)
  - Values: `id`, `created`, `updated`, `title`, `status`
- `--reverse` - Reverse sort order
- `--format` - Output format (default: `text`)
  - Values: `text`, `json`, `compact`

**Examples:**
```bash
# Default: active tasks
task list

# All tasks
task list --status all

# Blocked tasks only
task list --status blocked

# Sort by most recently updated
task list --status all --sort updated --reverse

# Sort alphabetically by title
task list --status all --sort title

# JSON output
task list --status all --format json

# Compact format (just ID and title)
task list --format compact
```

**Output (text format):**
```
#1    [active   ] Implement authentication
#2    [backlog  ] Write tests
#3    [blocked  ] Deploy to production
```

---

### show

Display full details for a task.

**Usage:**
```bash
task show <id>
```

**Arguments:**
- `id` (required) - Task ID number

**Description:**
Shows complete task information including:
- ID, title, status
- Created and updated timestamps
- Description
- Links to other tasks
- Tags
- Notes with timestamps and authors

**Examples:**
```bash
task show 42
```

**Output:**
```
Task #42: Implement authentication
Status: active
Created: 2025-11-03 10:30:00
Updated: 2025-11-03 14:20:00

Description:
Move from basic auth to OAuth2. Need to support Google and GitHub
providers.

Links:
  blocks #43
  parent #40 (Security improvements)

Tags: security, refactor

Notes:
  [2025-11-03 14:20] claude: Started implementation
  [2025-11-03 15:10] human: Don't forget to add tests
```

---

### update

Update task properties.

**Usage:**
```bash
task update <id> [--status STATUS] [--title TITLE] [--description DESC] [--note NOTE] [--author AUTHOR]
```

**Arguments:**
- `id` (required) - Task ID number

**Options:**
- `--status` - Change task status
  - Values: `backlog`, `active`, `blocked`, `done`, `cancelled`
- `--title` - Update task title
- `--description` - Update task description
- `--note` - Add a timestamped note
- `--author` - Note author name (default: `human`)

**Description:**
Updates one or more task properties. Multiple options can be combined in a single command.

**Examples:**
```bash
# Change status
task update 42 --status active

# Add a note
task update 42 --note "Started implementation"

# Add note with custom author
task update 42 --note "API endpoint complete" --author claude

# Update title
task update 42 --title "New title"

# Combine multiple updates
task update 42 --status done --note "Implementation complete"

# Update description
task update 42 --description "New detailed description"
```

**Output:**
```
Updated task #42
```

---

### link

Create relationships between tasks.

**Usage:**
```bash
task link <id> <target_id> [--type TYPE] [--label LABEL] [--bidirectional]
```

**Arguments:**
- `id` (required) - Source task ID
- `target_id` (required) - Target task ID

**Options:**
- `--type` - Link type (default: `relates_to`)
  - Values: `blocks`, `blocked_by`, `parent`, `child`, `relates_to`, `duplicates`
  - Custom types are also supported
- `--label` - Optional custom label for the link
- `--bidirectional` - Create reciprocal link automatically

**Description:**
Creates a directional link from one task to another. Links help establish task relationships like dependencies, hierarchies, and associations.

**Bidirectional Behavior:**
When using `--bidirectional`, the reciprocal link is automatically created:
- `blocks` ↔ `blocked_by`
- `parent` ↔ `child`
- Other types create the same type in reverse

**Examples:**
```bash
# Simple link
task link 5 4 --type blocked_by

# With label
task link 2 3 --type parent --label "Main feature"

# Bidirectional link (creates both directions)
task link 5 4 --type blocked_by --bidirectional
# Creates: 5 blocked_by 4, and 4 blocks 5

# Custom relationship type
task link 10 11 --type inspired_by
```

**Output:**
```
Linked task #5 to #4 (blocked_by)
Created reciprocal link: task #4 to #5 (blocks)
```

---

### unlink

Remove relationships between tasks.

**Usage:**
```bash
task unlink <id> <target_id> [--type TYPE] [--bidirectional]
```

**Arguments:**
- `id` (required) - Source task ID
- `target_id` (required) - Target task ID

**Options:**
- `--type` - Specific link type to remove (if empty, removes all links)
- `--bidirectional` - Remove reciprocal link as well

**Description:**
Removes links between tasks. If no type is specified, removes all links from source to target.

**Examples:**
```bash
# Remove specific link type
task unlink 5 4 --type blocked_by

# Remove all links between tasks
task unlink 5 4

# Remove bidirectional links
task unlink 5 4 --type blocked_by --bidirectional
```

**Output:**
```
Removed link (blocked_by) from task #5 to #4
Removed reciprocal link (blocks) from task #4 to #5
```

---

### search

Search tasks by keyword.

**Usage:**
```bash
task search <query> [--format FORMAT]
```

**Arguments:**
- `query` (required) - Search term

**Options:**
- `--format` - Output format (default: `text`)
  - Values: `text`, `json`, `compact`

**Description:**
Performs full-text search across:
- Task titles
- Descriptions
- Notes
- Tags

Search is case-insensitive.

**Examples:**
```bash
# Search for tasks
task search "authentication"

# Search with JSON output
task search "bug" --format json

# Search for specific terms
task search "OAuth"
```

**Output (text format):**
```
Found 3 task(s):

#42   [active   ] Implement authentication
      Move from basic auth to OAuth2. Need to support Google...

#43   [backlog  ] Add authentication tests

#44   [done     ] Research authentication providers
```

---

### context

Show project context optimized for LLM consumption.

**Usage:**
```bash
task context [--format FORMAT]
```

**Options:**
- `--format` - Output format (default: `text`)
  - Values: `text`, `json`

**Description:**
Provides a compact overview of the project's task status, designed to be included in LLM prompts after context compaction. Shows:
- Active tasks
- Recently completed tasks (last 7 days, up to 5 most recent)
- Summary statistics

**Examples:**
```bash
# Text format for humans
task context

# JSON format for machine consumption
task context --format json
```

**Output (text format):**
```
PROJECT CONTEXT

Active Tasks (3):
  #42   Implement authentication
  #43   Add rate limiting
  #44   Update documentation

Recently Completed (2):
  #40   Fix login bug (completed 2025-11-02)
  #41   Add logging (completed 2025-11-02)

Total: 15 tasks (3 active, 8 backlog, 4 done, 0 cancelled)
```

---

### serve

Start web UI server.

**Usage:**
```bash
task serve [--port PORT] [--no-browser]
```

**Options:**
- `--port` - Port number (default: `8080`)
- `--no-browser` - Don't automatically open browser

**Description:**
Starts a local HTTP server serving the web UI. Features include:
- **Board View**: Kanban board with columns for each status
- **List View**: Compact list of all tasks
- **Search**: Real-time filtering as you type
- **Task Details**: Click any task to see full information
- **Auto-refresh**: Updates every 5 seconds

**Examples:**
```bash
# Default (port 8080, auto-open browser)
task serve

# Custom port
task serve --port 3000

# Don't open browser automatically
task serve --no-browser
```

**Output:**
```
Starting task server on http://localhost:8080
Press Ctrl+C to stop
```

The server runs until stopped with Ctrl+C.

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Generic error (task not found, invalid input, etc.) |
| 2 | File system error (can't read/write files) |
| 3 | Not in a task repository (no `.tasks/` directory found) |

## Tips and Best Practices

### For Humans

1. **Use descriptive titles**: Keep them short but meaningful
2. **Add notes frequently**: Track progress with timestamped notes
3. **Link related tasks**: Use `blocks`/`blocked_by` for dependencies
4. **Search when needed**: Faster than listing and filtering
5. **Use the web UI**: Great for visualizing project status

### For LLMs

1. **Start with context**: Run `task context` after context compaction
2. **Create tasks early**: Don't use TODO comments, create tasks
3. **Update progress**: Add notes as you work
4. **Mark completion**: Update status to `done` when finished
5. **Search before creating**: Check if a similar task exists

### Integration Example

Add to `.clinerules` or Claude project instructions:

```markdown
## Task Management

After context compaction:
```bash
task context
```

Before starting work:
```bash
task list --status active
```

Create tasks:
```bash
task create "title" "description"
```

Update progress:
```bash
task update ID --note "progress note"
```

Mark complete:
```bash
task update ID --status done
```
```

## Performance

All commands are optimized for speed:
- `task list`: <5ms (uses cached index)
- `task create`: <5ms (writes one file + rebuilds index)
- `task show`: <3ms (single file read)
- `task update`: <5ms (update file + rebuild index)
- `task search`: <10ms (reads all task files)

The index is rebuilt on every create/update operation to ensure queries remain fast even with thousands of tasks.
