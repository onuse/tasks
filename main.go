package main

import (
	"fmt"
	"os"

	"github.com/onuse/tasks/internal/commands"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	var err error

	switch command {
	case "init":
		err = commands.Init()
	case "create":
		err = commands.Create(args)
	case "list":
		err = commands.List(args)
	case "show":
		err = commands.Show(args)
	case "update":
		err = commands.Update(args)
	case "link":
		err = commands.Link(args)
	case "unlink":
		err = commands.Unlink(args)
	case "tag":
		err = commands.Tag(args)
	case "untag":
		err = commands.Untag(args)
	case "merge":
		err = commands.Merge(args)
	case "search":
		err = commands.Search(args)
	case "context":
		err = commands.Context(args)
	case "serve":
		err = commands.Serve(args)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: task <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  init                           Initialize task tracking in current repository")
	fmt.Println("  create <title> [description]   Create a new task")
	fmt.Println("  list [--status STATUS]         List tasks (defaults to active)")
	fmt.Println("  show <id>                      Show full task details")
	fmt.Println("  update <id> [options]          Update a task")
	fmt.Println("  link <id> <target> [options]   Link two tasks together")
	fmt.Println("  unlink <id> <target> [options] Remove link between tasks")
	fmt.Println("  tag <id> <name>                Tag a task (creates label if needed)")
	fmt.Println("  untag <id> <name>              Remove a tag from a task")
	fmt.Println("  merge <source> <target>        Merge source task into target")
	fmt.Println("  search <query> [options]       Search tasks by keyword")
	fmt.Println("  context                        Show project context for LLMs")
	fmt.Println("  serve [--port PORT]            Start web UI server")
}
