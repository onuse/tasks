package commands

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/onuse/tasks/internal/store"
)

func Unlink(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: task unlink <id> <target_id> [--type TYPE] [--bidirectional]")
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

	// Parse flags
	fs := flag.NewFlagSet("unlink", flag.ExitOnError)
	linkType := fs.String("type", "", "Link type to remove (if empty, removes all links to target)")
	bidirectional := fs.Bool("bidirectional", false, "Remove reciprocal link as well")
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

	// Remove link
	if !sourceTask.RemoveLink(targetID, *linkType) {
		return fmt.Errorf("no link found from task #%d to #%d", sourceID, targetID)
	}

	sourceTask.Updated = time.Now()

	if err := s.WriteTask(sourceTask); err != nil {
		return err
	}

	// Rebuild index
	if err := s.RebuildIndex(); err != nil {
		return err
	}

	if *linkType != "" {
		fmt.Printf("Removed link (%s) from task #%d to #%d\n", *linkType, sourceID, targetID)
	} else {
		fmt.Printf("Removed all links from task #%d to #%d\n", sourceID, targetID)
	}

	// Handle bidirectional unlinking
	if *bidirectional {
		targetTask, err := s.ReadTask(targetID)
		if err != nil {
			return err
		}

		reciprocalType := ""
		if *linkType != "" {
			reciprocalType = getReciprocalLinkType(*linkType)
		}

		if targetTask.RemoveLink(sourceID, reciprocalType) {
			targetTask.Updated = time.Now()

			if err := s.WriteTask(targetTask); err != nil {
				return err
			}

			if err := s.RebuildIndex(); err != nil {
				return err
			}

			if reciprocalType != "" {
				fmt.Printf("Removed reciprocal link (%s) from task #%d to #%d\n", reciprocalType, targetID, sourceID)
			} else {
				fmt.Printf("Removed all reciprocal links from task #%d to #%d\n", targetID, sourceID)
			}
		}
	}

	return nil
}
