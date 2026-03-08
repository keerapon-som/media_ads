package entities

type SaveMediaResponse struct {
	MediaID   string `json:"media_id"`
	UploadURL string `json:"upload_url"`
}
