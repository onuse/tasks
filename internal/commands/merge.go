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

func Merge(args []string) error {
	fs := flag.NewFlagSet("merge", flag.ExitOnError)
	fs.Parse(args)

	if fs.NArg() < 2 {
		return fmt.Errorf("usage: task merge <source_id> <target_id>")
	}

	// Parse IDs
	sourceID, err := strconv.Atoi(fs.Arg(0))
	if err != nil {
		return fmt.Errorf("invalid source ID '%s'", fs.Arg(0))
	}

	targetID, err := strconv.Atoi(fs.Arg(1))
	if err != nil {
		return fmt.Errorf("invalid target ID '%s'", fs.Arg(1))
	}

	if sourceID == targetID {
		return fmt.Errorf("source and target cannot be the same task")
	}

	// Find task root
	rootDir, err := store.FindTaskRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run 'task init' to initialize task tracking in this repository.\n")
		os.Exit(3)
	}

	s := store.New(rootDir)

	// Read both tasks
	sourceTask, err := s.ReadTask(sourceID)
	if err != nil {
		return fmt.Errorf("source task #%d not found", sourceID)
	}

	targetTask, err := s.ReadTask(targetID)
	if err != nil {
		return fmt.Errorf("target task #%d not found", targetID)
	}

	// Read index to find all tasks
	index, err := s.ReadIndex()
	if err != nil {
		return err
	}

	tasksUpdated := 0

	// Update all tasks that link to source
	for _, entry := range index.Tasks {
		if entry.ID == sourceID || entry.ID == targetID {
			continue // Skip source and target themselves
		}

		t, err := s.ReadTask(entry.ID)
		if err != nil {
			continue
		}

		modified := false

		// Update links pointing to source
		for i := range t.Links {
			if t.Links[i].TargetID == sourceID {
				t.Links[i].TargetID = targetID
				modified = true
			}
		}

		if modified {
			t.Updated = time.Now()
			if err := s.WriteTask(t); err != nil {
				return err
			}
			tasksUpdated++
		}
	}

	// Merge source links into target
	for _, link := range sourceTask.Links {
		// Skip if target already has this link
		if !targetTask.HasLink(link.TargetID, link.Type) {
			targetTask.AddLink(link.TargetID, link.Type, link.Label)
		}
	}

	// Merge source notes into target
	for _, note := range sourceTask.Notes {
		targetTask.Notes = append(targetTask.Notes, note)
	}

	targetTask.Updated = time.Now()

	// Save target task
	if err := s.WriteTask(targetTask); err != nil {
		return err
	}

	// Cancel source task
	sourceTask.Status = task.StatusCancelled
	sourceTask.Updated = time.Now()
	sourceTask.Description = fmt.Sprintf("[MERGED INTO #%d] %s", targetID, sourceTask.Description)

	if err := s.WriteTask(sourceTask); err != nil {
		return err
	}

	// Rebuild index
	if err := s.RebuildIndex(); err != nil {
		return err
	}

	fmt.Printf("Merged task #%d into #%d\n", sourceID, targetID)
	fmt.Printf("Updated %d task(s) that referenced #%d\n", tasksUpdated, sourceID)
	fmt.Printf("Source task #%d marked as cancelled\n", sourceID)

	return nil
}
