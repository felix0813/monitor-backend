package models

import "time"

type TodoStatus string

const (
	TodoStatusPending TodoStatus = "pending"
	TodoStatusDone    TodoStatus = "done"
)

type TodoItem struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	Status      TodoStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TodoProject struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	OrderedIDs  []string             `json:"ordered_ids"`
	Items       map[string]*TodoItem `json:"items"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type TodoStore struct {
	OrderedIDs []string                `json:"ordered_ids"`
	Projects   map[string]*TodoProject `json:"projects"`
}
