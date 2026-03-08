package entities

import (
	"time"

	"github.com/google/uuid"
)

type EventHeader struct {
	ID             string    `json:"id"`
	PublishedAt    time.Time `json:"published_at"`
	IdempotencyKey string    `json:"idempotency_key"`
}

func NewEventHeader() EventHeader {
	return EventHeader{
		ID:             uuid.NewString(),
		PublishedAt:    time.Now().UTC(),
		IdempotencyKey: uuid.NewString(),
	}
}

type HelloCQRSEvent struct {
	Header EventHeader `json:"header"`
	Hello  string      `json:"hello"`
}

type ReceiveUploadCallbackEvent struct {
	Header      EventHeader `json:"header"`
	MediaID     string      `json:"media_id"`
	ObjectID    string      `json:"object_id"`
	ContentType string      `json:"content_type"`
	Success     bool        `json:"success"`
}

type RequestUploadCallbackEvent struct {
	Header      EventHeader `json:"header"`
	MediaID     string      `json:"media_id"`
	ObjectID    string      `json:"object_id"`
	ContentType string      `json:"content_type"`
	Success     bool        `json:"success"`
}
