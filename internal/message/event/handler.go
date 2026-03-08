package event

import (
	"media_ads/internal/domain"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type Handler struct {
	eventBus       *cqrs.EventBus
	commandBus     *cqrs.CommandBus
	mediaPublisher domain.MediaPublisherInterface
	objectLibrary  domain.ObjectLibraryInterface
	logger         watermill.LoggerAdapter
}

func NewHandler(
	eventBus *cqrs.EventBus,
	commandBus *cqrs.CommandBus,
	mediaPublisher domain.MediaPublisherInterface,
	objectLibrary domain.ObjectLibraryInterface,
	logger watermill.LoggerAdapter,
) Handler {

	if eventBus == nil {
		panic("eventBus is nil")
	}

	if commandBus == nil {
		panic("commandBus is nil")
	}

	if logger == nil {
		panic("Logger is nil")
	}

	handler := Handler{
		eventBus:       eventBus,
		commandBus:     commandBus,
		mediaPublisher: mediaPublisher,
		objectLibrary:  objectLibrary,
		logger:         logger,
	}

	return handler
}
