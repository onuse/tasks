package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/onuse/tasks/internal/store"
	"github.com/onuse/tasks/internal/task"
)

func List(args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	statusFlag := fs.String("status", "active", "Filter by status (backlog, next, active, blocked, done, cancelled, all)")
	formatFlag := fs.String("format", "text", "Output format (text, json, compact)")
	sortFlag := fs.String("sort", "id", "Sort by: id, created, updated, title, status")
	reverseFlag := fs.Bool("reverse", false, "Reverse sort order")
	fs.Parse(args)

	// Validate status
	filterStatus := *statusFlag
	if filterStatus != "all" && !task.IsValidStatus(filterStatus) {
		return fmt.Errorf("invalid status '%s' (must be: backlog, next, active, blocked, done, cancelled, all)", filterStatus)
	}

	// Find task root
	rootDir, err := store.FindTaskRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run 'task init' to initialize task tracking in this repository.\n")
		os.Exit(3)
	}

	s := store.New(rootDir)

	// Read index
	index, err := s.ReadIndex()
	if err != nil {
		return err
	}

	// Filter tasks
	var filtered []task.IndexEntry
	for _, entry := range index.Tasks {
		if filterStatus == "all" || string(entry.Status) == filterStatus {
			filtered = append(filtered, entry)
		}
	}

	// Sort tasks
	sortTasks(filtered, *sortFlag, *reverseFlag)

	// Output
	switch *formatFlag {
	case "json":
		return outputJSON(filtered)
	case "compact":
		return outputCompact(filtered)
	default:
		return outputText(filtered)
	}
}

func sortTasks(tasks []task.IndexEntry, sortBy string, reverse bool) {
	switch sortBy {
	case "created":
		sort.Slice(tasks, func(i, j int) bool {
			if reverse {
				return tasks[i].Created.After(tasks[j].Created)
			}
			return tasks[i].Created.Before(tasks[j].Created)
		})
	case "updated":
		sort.Slice(tasks, func(i, j int) bool {
			if reverse {
				return tasks[i].Updated.After(tasks[j].Updated)
			}
			return tasks[i].Updated.Before(tasks[j].Updated)
		})
	case "title":
		sort.Slice(tasks, func(i, j int) bool {
			if reverse {
				return tasks[i].Title > tasks[j].Title
			}
			return tasks[i].Title < tasks[j].Title
		})
	case "status":
		sort.Slice(tasks, func(i, j int) bool {
			if reverse {
				return tasks[i].Status > tasks[j].Status
			}
			return tasks[i].Status < tasks[j].Status
		})
	default: // "id"
		sort.Slice(tasks, func(i, j int) bool {
			if reverse {
				return tasks[i].ID > tasks[j].ID
			}
			return tasks[i].ID < tasks[j].ID
		})
	}
}

func outputText(tasks []task.IndexEntry) error {
	if len(tasks) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	for _, t := range tasks {
		fmt.Printf("#%-4d [%-9s] %s\n", t.ID, t.Status, t.Title)
	}
	return nil
}

func outputCompact(tasks []task.IndexEntry) error {
	if len(tasks) == 0 {
		return nil
	}

	for _, t := range tasks {
		fmt.Printf("#%d %s\n", t.ID, t.Title)
	}
	return nil
}

func outputJSON(tasks []task.IndexEntry) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
