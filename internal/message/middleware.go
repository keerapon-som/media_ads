package message

import (
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/lithammer/shortuuid/v3"
	"github.com/sirupsen/logrus"
)

func setCorrelationID(next message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {

		ctx := msg.Context()

		correlationID := msg.Metadata.Get("correlation_id")

		if correlationID == "" {
			correlationID = shortuuid.New()
		}

		ctx = log.ToContext(ctx, logrus.WithFields(logrus.Fields{"correlation_id": correlationID}))
		ctx = log.ContextWithCorrelationID(ctx, correlationID)

		msg.SetContext(ctx)

		return next(msg)
	}
}

func handlerMsgError(next message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		logger := log.FromContext(msg.Context())
		logger = logger.WithField("message_uuid", msg.UUID)

		topic := message.SubscribeTopicFromCtx(msg.Context())

		logger.Infof("Handling a message from topic [%s]", topic)

		msgs, err := next(msg)

		if err != nil {
			logger.WithFields(logrus.Fields{
				"error":        err.Error(),
				"message_uuid": msg.UUID,
			}).Info("Message handling error")

		}

		return msgs, err
	}
}
