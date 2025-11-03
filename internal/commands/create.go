package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/onuse/tasks/internal/store"
	"github.com/onuse/tasks/internal/task"
)

func Create(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: task create <title> [description]")
	}

	title := args[0]
	if title == "" {
		return fmt.Errorf("title cannot be empty")
	}

	description := ""
	if len(args) > 1 {
		description = args[1]
	}

	// Find task root
	rootDir, err := store.FindTaskRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run 'task init' to initialize task tracking in this repository.\n")
		os.Exit(3)
	}

	s := store.New(rootDir)

	// Read and update manifest
	manifest, err := s.ReadManifest()
	if err != nil {
		return err
	}

	taskID := manifest.NextID
	manifest.NextID++

	if err := s.WriteManifest(manifest); err != nil {
		return err
	}

	// Create task
	now := time.Now()
	newTask := task.Task{
		ID:           taskID,
		Created:      now,
		Updated:      now,
		Status:       task.StatusBacklog,
		Title:        title,
		Description:  description,
		Notes:        []task.Note{},
		Links:        []task.TaskLink{},
		Dependencies: []int{},
		Tags:         []string{},
	}

	if err := s.WriteTask(&newTask); err != nil {
		return err
	}

	// Rebuild index
	if err := s.RebuildIndex(); err != nil {
		return err
	}

	fmt.Printf("Created task #%d\n", taskID)
	return nil
}
