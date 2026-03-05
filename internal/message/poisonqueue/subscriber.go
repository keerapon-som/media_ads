package poisonqueue

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-sql/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/message"

	pg "database/sql"
)

func NewSubscriber(db *pg.DB, watermillLogger watermill.LoggerAdapter) (message.Subscriber, error) {

	return sql.NewSubscriber(db, sql.SubscriberConfig{
		// ConsumerGroup:    "svc-hello-cqrs-pg." + params.HandlerName,
		SchemaAdapter:    sql.DefaultPostgreSQLSchema{},
		OffsetsAdapter:   sql.DefaultPostgreSQLOffsetsAdapter{},
		InitializeSchema: true,
		// PollInterval:  5 * time.Second,
	}, watermillLogger)
}
