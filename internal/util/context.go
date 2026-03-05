package util

import (
	"context"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/lithammer/shortuuid"
	"github.com/sirupsen/logrus"
)

func NewContext(ctx context.Context) context.Context {
	correlationID := shortuuid.New()

	ctx = log.ToContext(ctx, logrus.WithFields(logrus.Fields{"correlation_id": correlationID}))
	ctx = log.ContextWithCorrelationID(ctx, correlationID)
	return ctx
}
