package db

import (
	"encoding/json"
	"errors"
	"monitor/models"
	"os"
	"path/filepath"
	"sync"
)

var (
	todosFile = filepath.Join("db", "todos.json")
	todosMu   sync.Mutex
)

func LoadTodoStore() (*models.TodoStore, error) {
	todosMu.Lock()
	defer todosMu.Unlock()

	return loadTodoStoreLocked()
}

func SaveTodoStore(store *models.TodoStore) error {
	todosMu.Lock()
	defer todosMu.Unlock()

	return saveTodoStoreLocked(store)
}

func loadTodoStoreLocked() (*models.TodoStore, error) {
	if err := ensureTodoFile(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(todosFile)
	if err != nil {
		return nil, err
	}

	store := &models.TodoStore{}
	if len(data) == 0 {
		return newTodoStore(), nil
	}

	if err := json.Unmarshal(data, store); err != nil {
		return nil, err
	}

	normalizeTodoStore(store)
	return store, nil
}

func saveTodoStoreLocked(store *models.TodoStore) error {
	if store == nil {
		return errors.New("todo store is nil")
	}

	normalizeTodoStore(store)

	if err := os.MkdirAll(filepath.Dir(todosFile), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(todosFile, data, 0o644)
}

func ensureTodoFile() error {
	if err := os.MkdirAll(filepath.Dir(todosFile), 0o755); err != nil {
		return err
	}

	if _, err := os.Stat(todosFile); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	return saveTodoStoreLocked(newTodoStore())
}

func newTodoStore() *models.TodoStore {
	return &models.TodoStore{
		OrderedIDs: make([]string, 0),
		Projects:   make(map[string]*models.TodoProject),
	}
}

func normalizeTodoStore(store *models.TodoStore) {
	if store.Projects == nil {
		store.Projects = make(map[string]*models.TodoProject)
	}
	if store.OrderedIDs == nil {
		store.OrderedIDs = make([]string, 0)
	}

	ordered := make([]string, 0, len(store.OrderedIDs))
	seen := make(map[string]struct{}, len(store.OrderedIDs))
	for _, id := range store.OrderedIDs {
		if id == "" {
			continue
		}
		project, exists := store.Projects[id]
		if !exists || project == nil {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		normalizeTodoProject(project)
		ordered = append(ordered, id)
		seen[id] = struct{}{}
	}

	for id, project := range store.Projects {
		if id == "" || project == nil {
			delete(store.Projects, id)
			continue
		}
		normalizeTodoProject(project)
		if _, exists := seen[id]; exists {
			continue
		}
		ordered = append(ordered, id)
		seen[id] = struct{}{}
	}

	store.OrderedIDs = ordered
}

func normalizeTodoProject(project *models.TodoProject) {
	if project.Items == nil {
		project.Items = make(map[string]*models.TodoItem)
	}
	if project.OrderedIDs == nil {
		project.OrderedIDs = make([]string, 0)
	}

	ordered := make([]string, 0, len(project.OrderedIDs))
	seen := make(map[string]struct{}, len(project.OrderedIDs))
	for _, id := range project.OrderedIDs {
		if id == "" {
			continue
		}
		item, exists := project.Items[id]
		if !exists || item == nil {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		normalizeTodoItem(item)
		ordered = append(ordered, id)
		seen[id] = struct{}{}
	}

	for id, item := range project.Items {
		if id == "" || item == nil {
			delete(project.Items, id)
			continue
		}
		normalizeTodoItem(item)
		if _, exists := seen[id]; exists {
			continue
		}
		ordered = append(ordered, id)
		seen[id] = struct{}{}
	}

	project.OrderedIDs = ordered
}

func normalizeTodoItem(item *models.TodoItem) {
	if item.Status == "" {
		item.Status = models.TodoStatusPending
	}
}
