# Task - Minimalist Task Management for LLM Workflows

A fast, local-first task management CLI optimized for LLM-driven development. Built with Go for speed and simplicity.

## Features

- **Fast**: Sub-10ms response time for all commands
- **Local-first**: All data stored in `.tasks/` directory, version controlled with git
- **LLM-native**: Designed for how LLMs work and think
- **Simple**: Single binary, no dependencies, works everywhere
- **Git-friendly**: One file per task minimizes merge conflicts

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
task list --status done
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

- `backlog` - Not started (default for new tasks)
- `active` - Currently being worked on
- `done` - Completed
- `cancelled` - Won't do

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

### Principles
- Use tasks instead of TODO comments in code
- Use tasks instead of creating TODO.md files
- Break large work into multiple tasks
- Keep task titles short and actionable
- Add notes as you make progress
```

## Error Handling

Exit codes:
- `0` - Success
- `1` - Generic error (task not found, invalid input)
- `2` - File system error
- `3` - Not in a task-enabled repository

## Performance

Tested on modern hardware with 1000+ tasks:
- `task list`: <5ms
- `task create`: <5ms
- `task show`: <3ms
- `task update`: <5ms
- `task context`: <5ms

## Development

### Building

```bash
go build -o task
```

### Running Tests

```bash
go test ./...
```

### Project Structure

```
tasks/
  main.go                 # CLI entry point
  internal/
    task/
      task.go            # Data structures
    store/
      store.go           # File I/O operations
    commands/
      init.go            # Command implementations
      create.go
      list.go
      show.go
      update.go
      context.go
```

## Philosophy

Do one thing well. Be the Git of task tracking for LLM workflows.

**This tool does NOT:**
- Sync across machines (use git)
- Support teams/permissions (it's local files)
- Integrate with Jira/Linear/etc
- Have a mobile app
- Track time automatically
- Support plugins
- Require configuration files

## License

MIT

## Contributing

Contributions welcome! Please open an issue or pull request.

## Roadmap

- [ ] Web UI (`task serve`)
- [ ] Full-text search
- [ ] Dependency tracking
- [ ] Export formats (CSV, Markdown)
- [ ] Git hooks integration
