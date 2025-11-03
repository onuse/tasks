package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/onuse/tasks/internal/task"
)

const (
	TasksDir     = ".tasks"
	TasksSubDir  = "tasks"
	ManifestFile = "manifest.json"
	IndexFile    = "index.json"
)

// Store handles all file I/O operations
type Store struct {
	rootDir string
}

// New creates a new Store instance
func New(rootDir string) *Store {
	return &Store{rootDir: rootDir}
}

// FindTaskRoot walks up the directory tree to find .tasks directory
func FindTaskRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		tasksDir := filepath.Join(dir, TasksDir)
		if _, err := os.Stat(tasksDir); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not in a task repository (no .tasks/ directory found)")
		}
		dir = parent
	}
}

// Init creates the .tasks directory structure
func (s *Store) Init() error {
	tasksPath := filepath.Join(s.rootDir, TasksDir)

	// Check if already exists
	if _, err := os.Stat(tasksPath); err == nil {
		return fmt.Errorf(".tasks/ directory already exists")
	}

	// Create directories
	if err := os.MkdirAll(filepath.Join(tasksPath, TasksSubDir), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Create manifest
	manifest := task.Manifest{
		NextID:  1,
		Created: time.Now(),
		Version: "1.0",
	}
	if err := s.WriteManifest(&manifest); err != nil {
		return err
	}

	// Create empty index
	index := task.Index{
		Tasks:   []task.IndexEntry{},
		Updated: time.Now(),
	}
	if err := s.WriteIndex(&index); err != nil {
		return err
	}

	return nil
}

// ReadManifest reads the manifest.json file
func (s *Store) ReadManifest() (*task.Manifest, error) {
	path := filepath.Join(s.rootDir, TasksDir, ManifestFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest task.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

// WriteManifest writes the manifest.json file atomically
func (s *Store) WriteManifest(manifest *task.Manifest) error {
	path := filepath.Join(s.rootDir, TasksDir, ManifestFile)
	return s.writeJSONAtomic(path, manifest)
}

// ReadIndex reads the index.json file
func (s *Store) ReadIndex() (*task.Index, error) {
	path := filepath.Join(s.rootDir, TasksDir, IndexFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read index: %w", err)
	}

	var index task.Index
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse index: %w", err)
	}

	return &index, nil
}

// WriteIndex writes the index.json file atomically
func (s *Store) WriteIndex(index *task.Index) error {
	path := filepath.Join(s.rootDir, TasksDir, IndexFile)
	return s.writeJSONAtomic(path, index)
}

// ReadTask reads a task file by ID
func (s *Store) ReadTask(id int) (*task.Task, error) {
	path := s.taskPath(id)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("task #%d not found", id)
		}
		return nil, fmt.Errorf("failed to read task: %w", err)
	}

	var t task.Task
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("failed to parse task: %w", err)
	}

	return &t, nil
}

// WriteTask writes a task file atomically
func (s *Store) WriteTask(t *task.Task) error {
	path := s.taskPath(t.ID)
	return s.writeJSONAtomic(path, t)
}

// RebuildIndex rebuilds the index from all task files
func (s *Store) RebuildIndex() error {
	tasksDir := filepath.Join(s.rootDir, TasksDir, TasksSubDir)
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return fmt.Errorf("failed to read tasks directory: %w", err)
	}

	var indexEntries []task.IndexEntry
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(tasksDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue // Skip files we can't read
		}

		var t task.Task
		if err := json.Unmarshal(data, &t); err != nil {
			continue // Skip files we can't parse
		}

		indexEntries = append(indexEntries, task.IndexEntry{
			ID:      t.ID,
			Status:  t.Status,
			Title:   t.Title,
			Created: t.Created,
			Updated: t.Updated,
		})
	}

	// Sort by ID
	sort.Slice(indexEntries, func(i, j int) bool {
		return indexEntries[i].ID < indexEntries[j].ID
	})

	index := task.Index{
		Tasks:   indexEntries,
		Updated: time.Now(),
	}

	return s.WriteIndex(&index)
}

// taskPath returns the file path for a task ID
func (s *Store) taskPath(id int) string {
	filename := fmt.Sprintf("%05d.json", id)
	return filepath.Join(s.rootDir, TasksDir, TasksSubDir, filename)
}

// writeJSONAtomic writes JSON data to a file atomically
func (s *Store) writeJSONAtomic(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to temp file first
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Rename to final path (atomic on most systems)
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath) // Clean up temp file
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}
