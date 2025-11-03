package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/onuse/tasks/internal/store"
	"github.com/onuse/tasks/internal/task"
)

func Search(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: task search <query> [--format FORMAT]")
	}

	query := strings.ToLower(args[0])

	// Parse flags
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	formatFlag := fs.String("format", "text", "Output format (text, json, compact)")
	fs.Parse(args[1:])

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

	// Search through all tasks
	var matches []task.Task
	for _, entry := range index.Tasks {
		// Read full task for searching
		t, err := s.ReadTask(entry.ID)
		if err != nil {
			continue // Skip tasks we can't read
		}

		// Check if query matches title, description, notes, or tags
		if matchesQuery(t, query) {
			matches = append(matches, *t)
		}
	}

	// Output results
	switch *formatFlag {
	case "json":
		return outputSearchJSON(matches)
	case "compact":
		return outputSearchCompact(matches)
	default:
		return outputSearchText(matches)
	}
}

func matchesQuery(t *task.Task, query string) bool {
	// Search in title
	if strings.Contains(strings.ToLower(t.Title), query) {
		return true
	}

	// Search in description
	if strings.Contains(strings.ToLower(t.Description), query) {
		return true
	}

	// Search in tags
	for _, tag := range t.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}

	// Search in notes
	for _, note := range t.Notes {
		if strings.Contains(strings.ToLower(note.Text), query) {
			return true
		}
	}

	return false
}

func outputSearchText(tasks []task.Task) error {
	if len(tasks) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	fmt.Printf("Found %d task(s):\n\n", len(tasks))
	for _, t := range tasks {
		fmt.Printf("#%-4d [%-9s] %s\n", t.ID, t.Status, t.Title)
		if t.Description != "" {
			// Show first 80 chars of description
			desc := t.Description
			if len(desc) > 80 {
				desc = desc[:77] + "..."
			}
			fmt.Printf("      %s\n", desc)
		}
		fmt.Println()
	}
	return nil
}

func outputSearchCompact(tasks []task.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	for _, t := range tasks {
		fmt.Printf("#%d %s\n", t.ID, t.Title)
	}
	return nil
}

func outputSearchJSON(tasks []task.Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
