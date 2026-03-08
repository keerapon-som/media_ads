package entities

import "os"

type MediaInfo struct {
	Filename    string
	Extension   string
	SizeBytes   int64
	ContentType string
	ProbeData   map[string]any
}

type DownloadResponse struct {
	Filename    string
	Extension   string
	SizeBytes   int64
	File        *os.File
	ContentType string
}

type CallbackUpdateUploadStatusRequest struct {
	MediaID     string `json:"media_id"`
	ObjectID    string `json:"object_id"`
	ContentType string `json:"content_type"`
	Success     bool   `json:"success"`
}
