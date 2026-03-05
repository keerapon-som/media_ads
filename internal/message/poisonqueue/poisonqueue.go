package poisonqueue

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
)

type Message struct {
	ID     string
	Reason string
}

type Handler struct {
	topicName  string
	subscriber message.Subscriber
	publisher  message.Publisher
}

func NewHandler(poisonQueueTopicName string, subscriber message.Subscriber, publisher message.Publisher) (*Handler, error) {
	return &Handler{
		topicName:  poisonQueueTopicName,
		subscriber: subscriber,
		publisher:  publisher,
	}, nil
}

func (h *Handler) Preview(ctx context.Context) ([]Message, error) {
	var result []Message
	err := h.iterate(ctx, func(msg *message.Message) (bool, error) {
		result = append(result, Message{
			ID:     msg.UUID,
			Reason: msg.Metadata.Get(middleware.ReasonForPoisonedKey),
		})
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (h *Handler) Remove(ctx context.Context, messageID string) error {
	var found bool
	err := h.iterate(ctx, func(msg *message.Message) (bool, error) {
		if msg.UUID == messageID {
			found = true
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		return err
	}

	if !found {
		return fmt.Errorf("message not found: %v", messageID)
	}

	return nil
}

func (h *Handler) Requeue(ctx context.Context, messageID string) error {
	var found bool
	err := h.iterate(ctx, func(msg *message.Message) (bool, error) {
		if msg.UUID == messageID {
			found = true

			originalTopic := msg.Metadata.Get(middleware.PoisonedTopicKey)

			err := h.publisher.Publish(originalTopic, msg)
			if err != nil {
				return false, err
			}

			return false, nil
		}

		return true, nil
	})
	if err != nil {
		return err
	}

	if !found {
		return fmt.Errorf("message not found: %v", messageID)
	}

	return nil
}

func (h *Handler) iterate(ctx context.Context, actionFunc func(msg *message.Message) (bool, error)) error {
	router, err := message.NewRouter(
		message.RouterConfig{},
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	firstMessageUUID := ""

	done := false

	router.AddHandler(
		"preview",
		h.topicName,
		h.subscriber,
		h.topicName,
		h.publisher,
		func(msg *message.Message) ([]*message.Message, error) {
			if done {
				cancel()
				return nil, errors.New("done")
			}

			if firstMessageUUID == "" {
				firstMessageUUID = msg.UUID
			} else if firstMessageUUID == msg.UUID {
				// we've read all messages
				done = true
				return nil, errors.New("done")
			}

			keep, err := actionFunc(msg)
			if err != nil {
				return nil, err
			}

			if !keep {
				if msg.UUID == firstMessageUUID {
					done = true
				}
				return nil, nil
			}

			return []*message.Message{msg}, nil
		},
	)

	err = router.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}
