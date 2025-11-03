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

func Link(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: task link <id> <target_id> [--type TYPE] [--label LABEL] [--bidirectional]")
	}

	// Parse source and target IDs
	sourceID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid source task ID '%s'", args[0])
	}

	targetID, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid target task ID '%s'", args[1])
	}

	if sourceID == targetID {
		return fmt.Errorf("cannot link a task to itself")
	}

	// Parse flags
	fs := flag.NewFlagSet("link", flag.ExitOnError)
	linkType := fs.String("type", task.LinkTypeRelatesTo, "Link type (blocks, blocked_by, parent, child, relates_to, duplicates)")
	label := fs.String("label", "", "Optional custom label for the link")
	bidirectional := fs.Bool("bidirectional", false, "Create reciprocal link")
	fs.Parse(args[2:])

	// Find task root
	rootDir, err := store.FindTaskRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run 'task init' to initialize task tracking in this repository.\n")
		os.Exit(3)
	}

	s := store.New(rootDir)

	// Read source task
	sourceTask, err := s.ReadTask(sourceID)
	if err != nil {
		return err
	}

	// Verify target task exists
	_, err = s.ReadTask(targetID)
	if err != nil {
		return err
	}

	// Add link
	sourceTask.AddLink(targetID, *linkType, *label)
	sourceTask.Updated = time.Now()

	if err := s.WriteTask(sourceTask); err != nil {
		return err
	}

	// Rebuild index
	if err := s.RebuildIndex(); err != nil {
		return err
	}

	fmt.Printf("Linked task #%d to #%d (%s)\n", sourceID, targetID, *linkType)

	// Handle bidirectional linking
	if *bidirectional {
		reciprocalType := getReciprocalLinkType(*linkType)

		targetTask, err := s.ReadTask(targetID)
		if err != nil {
			return err
		}

		targetTask.AddLink(sourceID, reciprocalType, *label)
		targetTask.Updated = time.Now()

		if err := s.WriteTask(targetTask); err != nil {
			return err
		}

		if err := s.RebuildIndex(); err != nil {
			return err
		}

		fmt.Printf("Created reciprocal link: task #%d to #%d (%s)\n", targetID, sourceID, reciprocalType)
	}

	return nil
}

// getReciprocalLinkType returns the reciprocal link type
func getReciprocalLinkType(linkType string) string {
	switch linkType {
	case task.LinkTypeBlocks:
		return task.LinkTypeBlockedBy
	case task.LinkTypeBlockedBy:
		return task.LinkTypeBlocks
	case task.LinkTypeParent:
		return task.LinkTypeChild
	case task.LinkTypeChild:
		return task.LinkTypeParent
	default:
		return linkType // For symmetric relationships like "relates_to"
	}
}
