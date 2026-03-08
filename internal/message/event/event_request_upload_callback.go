package event

import (
	"context"
	"log"
	"media_ads/internal/entities"
)

func (h *Handler) RequestUploadCallbackEvent(ctx context.Context, event *entities.RequestUploadCallbackEvent) error {

	log.Println("RequestUploadCallbackEvent received for MediaID:", event.MediaID)

	return h.objectLibrary.CallbackUpdateUploadSuccess(
		event.MediaID,
		event.ObjectID,
		event.ContentType,
		event.Success,
		"", // callbackURL is not needed in this case since the callback is triggered by the object library itself
	)
}
