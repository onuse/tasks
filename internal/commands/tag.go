package commands

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/onuse/tasks/internal/store"
	"github.com/onuse/tasks/internal/task"
)

func Tag(args []string) error {
	fs := flag.NewFlagSet("tag", flag.ExitOnError)
	fs.Parse(args)

	if fs.NArg() < 2 {
		return fmt.Errorf("usage: task tag <id> <tag_name>")
	}

	// Parse task ID
	taskID, err := strconv.Atoi(fs.Arg(0))
	if err != nil {
		return fmt.Errorf("invalid task ID '%s'", fs.Arg(0))
	}

	tagName := fs.Arg(1)

	// Find task root
	rootDir, err := store.FindTaskRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run 'task init' to initialize task tracking in this repository.\n")
		os.Exit(3)
	}

	s := store.New(rootDir)

	// Read the task
	t, err := s.ReadTask(taskID)
	if err != nil {
		return fmt.Errorf("task #%d not found", taskID)
	}

	// Find or create label task
	labelTask, err := findOrCreateLabel(s, tagName)
	if err != nil {
		return err
	}

	// Check if already tagged
	if t.HasLink(labelTask.ID, task.LinkTypeChild) {
		fmt.Printf("Task #%d already tagged with '%s'\n", taskID, tagName)
		return nil
	}

	// Add link from task to label (task is child of label)
	t.AddLink(labelTask.ID, task.LinkTypeChild, "")
	t.Updated = time.Now()

	// Save task
	if err := s.WriteTask(t); err != nil {
		return err
	}

	// Rebuild index
	if err := s.RebuildIndex(); err != nil {
		return err
	}

	fmt.Printf("Tagged task #%d with '%s' (label task #%d)\n", taskID, tagName, labelTask.ID)
	return nil
}

func Untag(args []string) error {
	fs := flag.NewFlagSet("untag", flag.ExitOnError)
	fs.Parse(args)

	if fs.NArg() < 2 {
		return fmt.Errorf("usage: task untag <id> <tag_name>")
	}

	// Parse task ID
	taskID, err := strconv.Atoi(fs.Arg(0))
	if err != nil {
		return fmt.Errorf("invalid task ID '%s'", fs.Arg(0))
	}

	tagName := fs.Arg(1)

	// Find task root
	rootDir, err := store.FindTaskRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run 'task init' to initialize task tracking in this repository.\n")
		os.Exit(3)
	}

	s := store.New(rootDir)

	// Read the task
	t, err := s.ReadTask(taskID)
	if err != nil {
		return fmt.Errorf("task #%d not found", taskID)
	}

	// Find label task
	labelTask, err := findLabelByName(s, tagName)
	if err != nil {
		return fmt.Errorf("label '%s' not found", tagName)
	}

	// Remove link
	if !t.RemoveLink(labelTask.ID, task.LinkTypeChild) {
		return fmt.Errorf("task #%d is not tagged with '%s'", taskID, tagName)
	}

	t.Updated = time.Now()

	// Save task
	if err := s.WriteTask(t); err != nil {
		return err
	}

	// Rebuild index
	if err := s.RebuildIndex(); err != nil {
		return err
	}

	fmt.Printf("Removed tag '%s' from task #%d\n", tagName, taskID)
	return nil
}

// findOrCreateLabel finds an existing label task by name (case-insensitive) or creates a new one
func findOrCreateLabel(s *store.Store, name string) (*task.Task, error) {
	// Try to find existing label
	labelTask, err := findLabelByName(s, name)
	if err == nil {
		return labelTask, nil
	}

	// Create new label task
	manifest, err := s.ReadManifest()
	if err != nil {
		return nil, err
	}

	newTask := &task.Task{
		ID:          manifest.NextID,
		Created:     time.Now(),
		Updated:     time.Now(),
		Status:      task.StatusLabel,
		Title:       name,
		Description: "",
		Notes:       []task.Note{},
		Links:       []task.TaskLink{},
		Tags:        []string{},
	}

	manifest.NextID++

	if err := s.WriteTask(newTask); err != nil {
		return nil, err
	}

	if err := s.WriteManifest(manifest); err != nil {
		return nil, err
	}

	if err := s.RebuildIndex(); err != nil {
		return nil, err
	}

	return newTask, nil
}

// findLabelByName finds a label task by name (case-insensitive)
func findLabelByName(s *store.Store, name string) (*task.Task, error) {
	index, err := s.ReadIndex()
	if err != nil {
		return nil, err
	}

	nameLower := strings.ToLower(name)

	// Look for label-status tasks with matching title
	for _, entry := range index.Tasks {
		if entry.Status == task.StatusLabel && strings.ToLower(entry.Title) == nameLower {
			return s.ReadTask(entry.ID)
		}
	}

	return nil, fmt.Errorf("label '%s' not found", name)
}
