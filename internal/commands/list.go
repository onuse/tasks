package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/onuse/tasks/internal/store"
	"github.com/onuse/tasks/internal/task"
)

func List(args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	statusFlag := fs.String("status", "active", "Filter by status (backlog, active, done, cancelled, all)")
	formatFlag := fs.String("format", "text", "Output format (text, json, compact)")
	fs.Parse(args)

	// Validate status
	filterStatus := *statusFlag
	if filterStatus != "all" && !task.IsValidStatus(filterStatus) {
		return fmt.Errorf("invalid status '%s' (must be: backlog, active, done, cancelled, all)", filterStatus)
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
