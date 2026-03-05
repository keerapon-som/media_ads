package command

import (
	"context"
	"fmt"
	"media_ads/internal/entities"
)

func (h *Handler) HelloCQRSCommand(ctx context.Context, command *entities.HelloCQRSCommand) error {

	fmt.Println("HelloCQRSCommand")
	return nil

}
