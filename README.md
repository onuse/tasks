# Task - Minimalist Task Management for LLM Workflows

A fast, local-first task management CLI optimized for LLM-driven development. Built with Go for speed and simplicity.

## Features

- **Fast**: Sub-10ms response time for all commands
- **Local-first**: All data stored in `.tasks/` directory, version controlled with git
- **LLM-native**: Designed for how LLMs work and think
- **Simple**: Single binary, no dependencies, works everywhere
- **Git-friendly**: One file per task minimizes merge conflicts
- **Task Relationships**: Link tasks with flexible relationship types (blocks, parent/child, etc.)
- **Full-text Search**: Search across titles, descriptions, notes, and tags
- **Web UI**: Beautiful Kanban board with real-time updates
- **Flexible**: No artificial constraints - all status transitions allowed

## Installation

### From Source

```bash
git clone https://github.com/onuse/tasks
cd tasks
go build -o task
```

Move the binary to your PATH (e.g., `~/.local/bin/` or `/usr/local/bin/`).

## Quick Start

### Initialize in a Repository

```bash
cd your-project
task init
git add .tasks && git commit -m "Initialize task tracking"
```

### Create Tasks

```bash
task create "Implement authentication" "Add OAuth2 support for Google and GitHub"
task create "Write tests"
task create "Update documentation"
```

### List Tasks

```bash
# List active tasks (default)
task list

# List all tasks
task list --status all

# List by status
task list --status backlog
task list --status next
task list --status blocked
task list --status done

# Sort tasks
task list --status all --sort updated --reverse  # Newest first
task list --status all --sort title              # Alphabetical
task list --status all --sort id --reverse       # By ID descending

# Available sorts: id, created, updated, title, status
```

### Update Tasks

```bash
# Change status
task update 1 --status active
task update 1 --status done

# Add notes
task update 1 --note "Started implementation"

# Update title or description
task update 1 --title "New title"
task update 1 --description "New description"

# Combine multiple updates
task update 1 --status active --note "Working on this now"
```

### View Task Details

```bash
task show 1
```

### Link Tasks

```bash
# Create a link between tasks
task link 5 4 --type blocked_by  # Task 5 is blocked by task 4

# With a custom label
task link 2 3 --type parent --label "Main feature"

# Create reciprocal links automatically
task link 5 4 --type blocked_by --bidirectional
# Creates: 5 blocked_by 4, and 4 blocks 5

# Remove a link
task unlink 2 3 --type blocks

# Link types: blocks, blocked_by, parent, child, relates_to, duplicates
```

### Search Tasks

```bash
# Search by keyword (searches title, description, notes, tags)
task search "authentication"

# With different output formats
task search "bug" --format json
task search "refactor" --format compact
```

### Web UI

```bash
# Start the web interface
task serve

# Custom port
task serve --port 3000

# Don't auto-open browser
task serve --no-browser
```

Features:
- Kanban board view with all 5 statuses
- List view for quick scanning
- Real-time search and filtering
- Click tasks to see full details
- Auto-refresh every 5 seconds

### Get Project Context

```bash
# For LLMs - shows active tasks and recent activity
task context

# JSON format for machine consumption
task context --format json
```

## File Structure

```
.tasks/
  manifest.json          # Next ID counter
  index.json            # Cached index for fast queries
  tasks/
    00001.json          # Individual task files
    00002.json
    00003.json
```

## Task Statuses

- `backlog` - Not started, future work (default for new tasks)
- `next` - Prioritized and ready to work on
- `active` - Currently being worked on
- `blocked` - Waiting on dependencies or external factors
- `done` - Completed
- `cancelled` - Won't do

All status transitions are allowed - you can move tasks between any statuses freely.

The typical workflow is: `backlog` → `next` → `active` → `done`

## Output Formats

### Text (Default)

```
#1    [active   ] Implement authentication
#2    [backlog  ] Write tests
```

### JSON

```bash
task list --format json
```

```json
[
  {
    "id": 1,
    "status": "active",
    "title": "Implement authentication",
    "created": "2025-11-03T10:30:00Z",
    "updated": "2025-11-03T14:20:00Z"
  }
]
```

### Compact

```bash
task list --format compact
```

```
#1 Implement authentication
#2 Write tests
```

## LLM Integration

Add this to your `.clinerules` or Claude project instructions:

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