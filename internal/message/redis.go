package message

import (
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

func NewRedisStreamPublisher(redisClient *redis.Client, watermillLogger watermill.LoggerAdapter) message.Publisher {

	redisPublisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client:     redisClient,
		Marshaller: redisstream.DefaultMarshallerUnmarshaller{},
	}, watermillLogger)
	if err != nil {
		panic(err)
	}

	publisher := log.CorrelationPublisherDecorator{
		Publisher: redisPublisher,
	}

	return publisher

}
