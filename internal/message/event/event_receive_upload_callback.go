package event

import (
	"context"
	"log"
	"media_ads/internal/entities"
)

func (h *Handler) ReceiveUploadCallbackEvent(ctx context.Context, event *entities.ReceiveUploadCallbackEvent) error {

	log.Println("ReceiveUploadCallbackEvent received for MediaID:", event.MediaID)

	if event.Success {
		err := h.mediaPublisher.UpdateMediaUploadCompleted(event.MediaID, event.ObjectID, event.ContentType)
		if err != nil {
			h.logger.Error("Error updating media upload completed", err, nil)
			return err
		}
	} else {
		err := h.mediaPublisher.UpdateMediaUploadFailed(event.MediaID)
		if err != nil {
			h.logger.Error("Error updating media upload failed", err, nil)
			return err
		}
	}

	return nil
}
