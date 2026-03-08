package entities

import "time"

type ObjectLibraryRepo struct {
	ObjectID    string         `json:"object_id"`
	Key         string         `json:"key"`
	Filename    string         `json:"filename"`
	Extension   string         `json:"extension"`
	SizeBytes   int64          `json:"size_bytes"`
	ContentType string         `json:"content_type"`
	ProbeData   map[string]any `json:"probe_data"`
	IsPublished bool           `json:"is_published"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type UploadSlotRepo struct {
	UploadID  string    `json:"upload_id"`
	Status    string    `json:"status"` // pending, completed, failed
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
