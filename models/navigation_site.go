package models

import "time"

type NavigationSite struct {
	ID          string    `json:"id"`
	ImageURL    string    `json:"image_url"`
	URL         string    `json:"url"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type NavigationSiteStore struct {
	OrderedIDs []string                   `json:"ordered_ids"`
	Sites      map[string]*NavigationSite `json:"sites"`
}
