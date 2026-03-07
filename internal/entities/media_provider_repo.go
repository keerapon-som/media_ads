package entities

type MediaArchiveRepo struct {
	ObjectID    string         `json:"object_id"`
	Key         string         `json:"key"`
	Filename    string         `json:"filename"`
	Extension   string         `json:"extension"`
	SizeBytes   int64          `json:"size_bytes"`
	ContentType string         `json:"content_type"`
	ProbeData   map[string]any `json:"probe_data"`
}
