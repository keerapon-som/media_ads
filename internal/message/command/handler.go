package command

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type Handler struct {
	eventBus   *cqrs.EventBus
	commandBus *cqrs.CommandBus

	logger watermill.LoggerAdapter
}

func NewHandler(
	eventBus *cqrs.EventBus,
	commandBus *cqrs.CommandBus,
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
		eventBus:   eventBus,
		commandBus: commandBus,
		logger:     logger,
	}

	return handler
}
