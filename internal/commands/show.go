package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/onuse/tasks/internal/store"
)

func Show(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: task show <id>")
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid task ID '%s'", args[0])
	}

	// Find task root
	rootDir, err := store.FindTaskRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run 'task init' to initialize task tracking in this repository.\n")
		os.Exit(3)
	}

	s := store.New(rootDir)

	// Read task
	t, err := s.ReadTask(id)
	if err != nil {
		return err
	}

	// Display task
	fmt.Printf("Task #%d: %s\n", t.ID, t.Title)
	fmt.Printf("Status: %s\n", t.Status)
	fmt.Printf("Created: %s\n", t.Created.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", t.Updated.Format("2006-01-02 15:04:05"))
	fmt.Println()

	if t.Description != "" {
		fmt.Println("Description:")
		fmt.Println(t.Description)
		fmt.Println()
	}

	if len(t.Links) > 0 {
		fmt.Println("Links:")
		for _, link := range t.Links {
			if link.Label != "" {
				fmt.Printf("  %s #%d (%s)\n", link.Type, link.TargetID, link.Label)
			} else {
				fmt.Printf("  %s #%d\n", link.Type, link.TargetID)
			}
		}
		fmt.Println()
	}

	if len(t.Dependencies) > 0 {
		fmt.Print("Dependencies (deprecated): ")
		deps := make([]string, len(t.Dependencies))
		for i, dep := range t.Dependencies {
			deps[i] = fmt.Sprintf("#%d", dep)
		}
		fmt.Println(strings.Join(deps, ", "))
		fmt.Println()
	}

	if len(t.Tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(t.Tags, ", "))
		fmt.Println()
	}

	if len(t.Notes) > 0 {
		fmt.Println("Notes:")
		for _, note := range t.Notes {
			fmt.Printf("  [%s] %s: %s\n",
				note.Timestamp.Format("2006-01-02 15:04"),
				note.Author,
				note.Text)
		}
	}

	return nil
}
