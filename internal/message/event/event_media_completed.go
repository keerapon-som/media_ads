package event

import (
	"context"
	"fmt"
	"media_ads/internal/entities"
)

func (h *Handler) HelloCQRSEvent(ctx context.Context, event *entities.HelloCQRSEvent) error {
	fmt.Println("HelloCQRSEvent")

	return nil
}
