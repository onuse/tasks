package task

import (
	"time"
)

// Status represents the current state of a task
type Status string

const (
	StatusBacklog   Status = "backlog"
	StatusNext      Status = "next"
	StatusActive    Status = "active"
	StatusBlocked   Status = "blocked"
	StatusDone      Status = "done"
	StatusCancelled Status = "cancelled"
	StatusLabel     Status = "label"
)

// ValidStatuses returns all valid status values
func ValidStatuses() []Status {
	return []Status{StatusBacklog, StatusNext, StatusActive, StatusBlocked, StatusDone, StatusCancelled, StatusLabel}
}

// IsValidStatus checks if a status string is valid
func IsValidStatus(s string) bool {
	status := Status(s)
	for _, valid := range ValidStatuses() {
		if status == valid {
			return true
		}
	}
	return false
}

// Note represents a timestamped note on a task
type Note struct {
	Timestamp time.Time `json:"timestamp"`
	Author    string    `json:"author"`
	Text      string    `json:"text"`
}

// TaskLink represents a relationship between tasks
type TaskLink struct {
	TargetID int    `json:"target_id"`           // ID of the linked task
	Type     string `json:"type"`                // Link type: "blocks", "blocked_by", "parent", "child", "relates_to", etc.
	Label    string `json:"label,omitempty"`     // Optional custom label for the link
}

// Common link types
const (
	LinkTypeBlocks     = "blocks"
	LinkTypeBlockedBy  = "blocked_by"
	LinkTypeParent     = "parent"
	LinkTypeChild      = "child"
	LinkTypeRelatesTo  = "relates_to"
	LinkTypeDuplicates = "duplicates"
)

// Task represents a single task
type Task struct {
	ID           int        `json:"id"`
	Created      time.Time  `json:"created"`
	Updated      time.Time  `json:"updated"`
	Status       Status     `json:"status"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Notes        []Note     `json:"notes"`
	Links        []TaskLink `json:"links"`
	Dependencies []int      `json:"dependencies"` // Deprecated: kept for backward compatibility, use Links instead
	Tags         []string   `json:"tags"`
}

// IndexEntry represents a minimal task entry for fast queries
type IndexEntry struct {
	ID      int       `json:"id"`
	Status  Status    `json:"status"`
	Title   string    `json:"title"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

// Index represents the cached index of all tasks
type Index struct {
	Tasks   []IndexEntry `json:"tasks"`
	Updated time.Time    `json:"updated"`
}

// Manifest represents the manifest.json file
type Manifest struct {
	NextID  int       `json:"next_id"`
	Created time.Time `json:"created"`
	Version string    `json:"version"`
}

// Helper methods for Task

// AddLink adds a link to another task
func (t *Task) AddLink(targetID int, linkType string, label string) {
	// Check if link already exists
	for _, link := range t.Links {
		if link.TargetID == targetID && link.Type == linkType {
			return // Link already exists
		}
	}

	t.Links = append(t.Links, TaskLink{
		TargetID: targetID,
		Type:     linkType,
		Label:    label,
	})
}

// RemoveLink removes a link to another task
func (t *Task) RemoveLink(targetID int, linkType string) bool {
	for i, link := range t.Links {
		if link.TargetID == targetID && (linkType == "" || link.Type == linkType) {
			t.Links = append(t.Links[:i], t.Links[i+1:]...)
			return true
		}
	}
	return false
}

// GetLinks returns all links of a specific type
func (t *Task) GetLinks(linkType string) []TaskLink {
	var result []TaskLink
	for _, link := range t.Links {
		if linkType == "" || link.Type == linkType {
			result = append(result, link)
		}
	}
	return result
}

// HasLink checks if a link exists
func (t *Task) HasLink(targetID int, linkType string) bool {
	for _, link := range t.Links {
		if link.TargetID == targetID && (linkType == "" || link.Type == linkType) {
			return true
		}
	}
	return false
}
