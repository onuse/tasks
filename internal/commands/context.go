package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/onuse/tasks/internal/store"
	"github.com/onuse/tasks/internal/task"
)

type ContextOutput struct {
	Next               []ContextTask `json:"next"`
	Active             []ContextTask `json:"active"`
	RecentlyCompleted  []ContextTask `json:"recently_completed"`
	Summary            Summary       `json:"summary"`
}

type ContextTask struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed string `json:"completed,omitempty"`
}

type Summary struct {
	Total     int `json:"total"`
	Next      int `json:"next"`
	Active    int `json:"active"`
	Backlog   int `json:"backlog"`
	Done      int `json:"done"`
	Cancelled int `json:"cancelled"`
}

func Context(args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("context", flag.ExitOnError)
	formatFlag := fs.String("format", "text", "Output format (text, json)")
	fs.Parse(args)

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

	// Organize tasks
	var next []task.IndexEntry
	var active []task.IndexEntry
	var completed []task.IndexEntry
	summary := Summary{}

	for _, entry := range index.Tasks {
		summary.Total++

		switch entry.Status {
		case task.StatusNext:
			summary.Next++
			next = append(next, entry)
		case task.StatusActive:
			summary.Active++
			active = append(active, entry)
		case task.StatusBacklog:
			summary.Backlog++
		case task.StatusDone:
			summary.Done++
			completed = append(completed, entry)
		case task.StatusCancelled:
			summary.Cancelled++
		}
	}

	// Get recently completed (last 7 days)
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	var recentCompleted []task.IndexEntry
	for _, entry := range completed {
		if entry.Updated.After(sevenDaysAgo) {
			recentCompleted = append(recentCompleted, entry)
		}
	}

	// Limit to most recent 5
	if len(recentCompleted) > 5 {
		recentCompleted = recentCompleted[len(recentCompleted)-5:]
	}

	// Output
	switch *formatFlag {
	case "json":
		return outputContextJSON(next, active, recentCompleted, summary)
	default:
		return outputContextText(next, active, recentCompleted, summary)
	}
}

func outputContextText(next, active, completed []task.IndexEntry, summary Summary) error {
	fmt.Println("PROJECT CONTEXT")
	fmt.Println()

	if len(next) > 0 {
		fmt.Printf("Next Tasks (%d):\n", len(next))
		for _, t := range next {
			fmt.Printf("  #%-4d %s\n", t.ID, t.Title)
		}
		fmt.Println()
	}

	if len(active) > 0 {
		fmt.Printf("Active Tasks (%d):\n", len(active))
		for _, t := range active {
			fmt.Printf("  #%-4d %s\n", t.ID, t.Title)
		}
		fmt.Println()
	} else if len(next) == 0 {
		fmt.Println("Active Tasks: None")
		fmt.Println()
	}

	if len(completed) > 0 {
		fmt.Printf("Recently Completed (%d):\n", len(completed))
		for _, t := range completed {
			fmt.Printf("  #%-4d %s (completed %s)\n",
				t.ID, t.Title, t.Updated.Format("2006-01-02"))
		}
		fmt.Println()
	}

	fmt.Printf("Total: %d tasks (%d next, %d active, %d backlog, %d done, %d cancelled)\n",
		summary.Total, summary.Next, summary.Active, summary.Backlog, summary.Done, summary.Cancelled)

	return nil
}

func outputContextJSON(next, active, completed []task.IndexEntry, summary Summary) error {
	output := ContextOutput{
		Next:              make([]ContextTask, len(next)),
		Active:            make([]ContextTask, len(active)),
		RecentlyCompleted: make([]ContextTask, len(completed)),
		Summary:           summary,
	}

	for i, t := range next {
		output.Next[i] = ContextTask{
			ID:    t.ID,
			Title: t.Title,
		}
	}

	for i, t := range active {
		output.Active[i] = ContextTask{
			ID:    t.ID,
			Title: t.Title,
		}
	}

	for i, t := range completed {
		output.RecentlyCompleted[i] = ContextTask{
			ID:        t.ID,
			Title:     t.Title,
			Completed: t.Updated.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
