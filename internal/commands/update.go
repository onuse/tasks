package commands

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/onuse/tasks/internal/store"
	"github.com/onuse/tasks/internal/task"
)

func Update(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: task update <id> [--status STATUS] [--note NOTE] [--title TITLE] [--description DESC]")
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid task ID '%s'", args[0])
	}

	// Parse flags
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	statusFlag := fs.String("status", "", "New status")
	noteFlag := fs.String("note", "", "Add a note")
	titleFlag := fs.String("title", "", "New title")
	descFlag := fs.String("description", "", "New description")
	authorFlag := fs.String("author", "human", "Note author")
	fs.Parse(args[1:])

	// Validate status if provided
	if *statusFlag != "" && !task.IsValidStatus(*statusFlag) {
		return fmt.Errorf("invalid status '%s' (must be: backlog, active, done, cancelled)", *statusFlag)
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

	// Apply updates
	updated := false

	if *statusFlag != "" {
		t.Status = task.Status(*statusFlag)
		updated = true
	}

	if *titleFlag != "" {
		t.Title = *titleFlag
		updated = true
	}

	if *descFlag != "" {
		t.Description = *descFlag
		updated = true
	}

	if *noteFlag != "" {
		note := task.Note{
			Timestamp: time.Now(),
			Author:    *authorFlag,
			Text:      *noteFlag,
		}
		t.Notes = append(t.Notes, note)
		updated = true
	}

	if !updated {
		return fmt.Errorf("no updates specified")
	}

	// Update timestamp
	t.Updated = time.Now()

	// Write task
	if err := s.WriteTask(t); err != nil {
		return err
	}

	// Rebuild index
	if err := s.RebuildIndex(); err != nil {
		return err
	}

	fmt.Printf("Updated task #%d\n", id)
	return nil
}
