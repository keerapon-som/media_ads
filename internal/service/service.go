package service

import (
	"context"
	"errors"
	"media_ads/internal/config"
	"media_ads/internal/domain"
	"media_ads/internal/message"
	"media_ads/internal/message/command"
	"media_ads/internal/message/event"
	"net/http"
	"sync"

	httpCatchup "media_ads/internal/http"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	watermillMessage "github.com/ThreeDotsLabs/watermill/message"
)

type Service struct {
	listMessageRouter []*watermillMessage.Router
	watermillLogger   *log.WatermillLogrusAdapter
	fiberApp          *fiber.App
}

func New(redisClient *redis.Client,
	internalSecureToken string,
	mediaPublisher *domain.MediaPublisher,
	mediaProvider domain.ObjectLibraryInterface,
	numberConcurrent int,
) *Service {

	watermillLogger := log.NewWatermill(logrus.NewEntry(logrus.StandardLogger()))

	var publisher watermillMessage.Publisher
	publisher = message.NewRedisStreamPublisher(redisClient, watermillLogger)

	eventBus := event.NewBus(publisher)
	commandBus := command.NewBus(publisher)

	eventHandler := event.NewHandler(eventBus, commandBus, watermillLogger)
	eventProcessorConfig := event.NewProcessorConfig(redisClient, watermillLogger)

	commandHandler := command.NewHandler(eventBus, commandBus, watermillLogger)
	commandProcessorConfig := command.NewProcessorConfig(redisClient, watermillLogger)

	listRouter := []*watermillMessage.Router{}

	for i := 0; i < numberConcurrent; i++ {
		watermillroute := message.NewWatermillRouter(publisher, eventProcessorConfig, &eventHandler, commandProcessorConfig, &commandHandler, watermillLogger)
		listRouter = append(listRouter, watermillroute.Router())
	}

	mediaProviderHTTPHandler := httpCatchup.NewObjectLibraryProviderHTTPHandler(mediaProvider, commandBus, eventBus, watermillLogger)

	return &Service{
		watermillLogger:   watermillLogger,
		listMessageRouter: listRouter,
		fiberApp:          httpCatchup.NewHTTPRouter(mediaProviderHTTPHandler, internalSecureToken),
	}

}

func (s *Service) Run(ctx context.Context) error {

	errgroup, ctx := errgroup.WithContext(ctx)

	wg := sync.WaitGroup{}

	wg.Add(len(s.listMessageRouter))

	for num, router := range s.listMessageRouter {

		var err error
		go func() {
			err = router.Run(ctx)
		}()

		if err != nil {
			return err
		}

		<-router.Running()
		wg.Done()

		s.watermillLogger.Log.Infof("Run Router %d", num+1)

	}

	wg.Wait()

	errgroup.Go(func() error {

		err := s.fiberApp.Listen(":" + config.GetConfig().ServerConfig.HTTP.Port)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.watermillLogger.Log.Errorf("Error starting HTTP server: %v", err)
			return err
		}

		return nil
	})

	errgroup.Go(func() error {
		<-ctx.Done()

		s.watermillLogger.Log.Info("Got signal to shutdown")
		return s.fiberApp.Shutdown()
	})

	return errgroup.Wait()
}
