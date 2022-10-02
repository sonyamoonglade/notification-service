package app_middlewares

import (
	"github.com/sonyamoonglade/notification-service/internal/events"
	"github.com/sonyamoonglade/notification-service/internal/events/middlewares"
	"go.uber.org/zap"
)

type AppMiddlewares struct {
	*event_middlewares.DoesExist
}

func New(logger *zap.SugaredLogger, eventService events.Service) *AppMiddlewares {
	return &AppMiddlewares{
		event_middlewares.NewDoesExist(logger, eventService),
	}
}
