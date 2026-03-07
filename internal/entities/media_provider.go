package entities

type MediaInfo struct {
	Filename    string
	Extension   string
	SizeBytes   int64
	ContentType string
	ProbeData   map[string]any
}
