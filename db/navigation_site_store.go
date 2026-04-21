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
	navigationSitesFile = filepath.Join("db", "navigation_sites.json")
	navigationSitesMu   sync.Mutex
)

func LoadNavigationSiteStore() (*models.NavigationSiteStore, error) {
	navigationSitesMu.Lock()
	defer navigationSitesMu.Unlock()

	return loadNavigationSiteStoreLocked()
}

func SaveNavigationSiteStore(store *models.NavigationSiteStore) error {
	navigationSitesMu.Lock()
	defer navigationSitesMu.Unlock()

	return saveNavigationSiteStoreLocked(store)
}

func loadNavigationSiteStoreLocked() (*models.NavigationSiteStore, error) {
	if err := ensureNavigationSiteFile(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(navigationSitesFile)
	if err != nil {
		return nil, err
	}

	store := &models.NavigationSiteStore{}
	if len(data) == 0 {
		return newNavigationSiteStore(), nil
	}

	if err := json.Unmarshal(data, store); err != nil {
		return nil, err
	}

	normalizeNavigationSiteStore(store)
	return store, nil
}

func saveNavigationSiteStoreLocked(store *models.NavigationSiteStore) error {
	if store == nil {
		return errors.New("navigation site store is nil")
	}

	normalizeNavigationSiteStore(store)

	if err := os.MkdirAll(filepath.Dir(navigationSitesFile), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(navigationSitesFile, data, 0o644)
}

func ensureNavigationSiteFile() error {
	if err := os.MkdirAll(filepath.Dir(navigationSitesFile), 0o755); err != nil {
		return err
	}

	if _, err := os.Stat(navigationSitesFile); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	return saveNavigationSiteStoreLocked(newNavigationSiteStore())
}

func newNavigationSiteStore() *models.NavigationSiteStore {
	return &models.NavigationSiteStore{
		OrderedIDs: make([]string, 0),
		Sites:      make(map[string]*models.NavigationSite),
	}
}

func normalizeNavigationSiteStore(store *models.NavigationSiteStore) {
	if store.Sites == nil {
		store.Sites = make(map[string]*models.NavigationSite)
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
		if _, exists := store.Sites[id]; !exists {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		ordered = append(ordered, id)
		seen[id] = struct{}{}
	}

	for id := range store.Sites {
		if id == "" {
			delete(store.Sites, id)
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		ordered = append(ordered, id)
		seen[id] = struct{}{}
	}

	store.OrderedIDs = ordered
}
