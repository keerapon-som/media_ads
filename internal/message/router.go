package message

import (
	"media_ads/internal/message/command"
	"media_ads/internal/message/event"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
)

const (
	POISONQUEUE_TOPIC_NAME = "poison_queue"
)

type WatermillRouter struct {
	router           *message.Router
	eventProcessor   *cqrs.EventProcessor
	commandProcessor *cqrs.CommandProcessor
}

func NewWatermillRouter(
	publisher message.Publisher,
	eventProcessorConfig cqrs.EventProcessorConfig,
	eventHandler *event.Handler,
	CommandProcessorConfig cqrs.CommandProcessorConfig,
	commandHandler *command.Handler,
	watermillLogger watermill.LoggerAdapter) *WatermillRouter {

	router, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		panic(err)
	}

	router.AddMiddleware(setCorrelationID, handlerMsgError, middleware.Retry{
		MaxRetries:      10,
		InitialInterval: time.Second,
		MaxInterval:     time.Minute * 15,
		Multiplier:      2,
		Logger:          watermillLogger,
	}.Middleware)

	eventProcessor, err := cqrs.NewEventProcessorWithConfig(router, eventProcessorConfig)
	if err != nil {
		panic(err)
	}

	eventProcessor.AddHandlers(
		cqrs.NewEventHandler(
			"HelloCQRSEvent",
			eventHandler.HelloCQRSEvent,
		),
	)

	commandProcessor, err := cqrs.NewCommandProcessorWithConfig(router, CommandProcessorConfig)
	if err != nil {
		panic(err)
	}

	commandProcessor.AddHandlers(
		cqrs.NewCommandHandler(
			"HelloCQRSCommand",
			commandHandler.HelloCQRSCommand,
		),
	)

	return &WatermillRouter{
		router:           router,
		eventProcessor:   eventProcessor,
		commandProcessor: commandProcessor,
	}

}

func (w *WatermillRouter) Router() *message.Router {
	return w.router
}

func (w *WatermillRouter) GetEventSubscribeTopicName() []string {
	listTopic := []string{}
	for _, topic := range w.eventProcessor.Handlers() {
		listTopic = append(listTopic, "events_"+topic.HandlerName())
	}
	return listTopic
}

func (w *WatermillRouter) GetCommandSubscribeTopicName() []string {
	listTopic := []string{}
	for _, topic := range w.commandProcessor.Handlers() {
		listTopic = append(listTopic, "commands_"+topic.HandlerName())
	}
	return listTopic
}
