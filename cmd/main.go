package main

import (
	"context"
	"errors"
	"github.com/joho/godotenv"
	"github.com/sonyamoonglade/delivery-service/pkg/logging"
	"github.com/sonyamoonglade/notification-service/config"
	"github.com/sonyamoonglade/notification-service/internal/events"
	"github.com/sonyamoonglade/notification-service/internal/events/middleware"
	"github.com/sonyamoonglade/notification-service/internal/subscription"
	"github.com/sonyamoonglade/notification-service/pkg/postgres"
	"github.com/sonyamoonglade/notification-service/pkg/server"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	log.Println("booting an application")

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	logger, err := logging.WithCfg(&logging.Config{
		Level:    zap.NewAtomicLevelAt(zap.DebugLevel),
		DevMode:  true,
		Encoding: logging.JSON,
	})
	if err != nil {
		log.Fatalf("could not get zap logger. %s", err.Error())
	}

	//Make sure to load env variables on local development
	if err := godotenv.Load(".env.local"); err != nil {
		logger.Fatalf("could not load .env.local %s", err.Error())
	}

	appCfg, err := config.GetAppConfig()
	if err != nil {
		logger.Fatalf("could not get app config. %s", err.Error())
	}
	logger.Info("initialized config")
	if appCfg.Env == "development" {
		logger.Info("environment variables are loaded (development only)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	pg, err := postgres.New(ctx, appCfg.DatabaseURL)
	if err != nil {
		logger.Fatalf("could not create pool. %s", err.Error())
	}
	logger.Info("database has connected")

	srv, router := server.NewServer(&appCfg)

	eventsStorage := events.NewEventStorage(logger, pg.Pool)
	eventsService := events.NewEventsService(logger, eventsStorage)
	eventsMiddleware := middleware.NewEventsMiddlewares(logger, eventsService)

	subscriptionStorage := subscription.NewSubscriptionStorage(logger, pg.Pool)
	subscriptionService := subscription.NewSubscriptionService(logger, subscriptionStorage)
	subscriptionTransport := subscription.NewSubscriptionTransport(logger, subscriptionService, eventsMiddleware, eventsService)

	subscriptionTransport.InitRoutes(router)
	logger.Info("initialized routes")

	if err = eventsService.ReadEvents(ctx); err != nil {
		logger.Fatalf("could not load base events. %s", err.Error())
	}
	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("could not start server. %s", err.Error())
		}
	}()
	logger.Infof("server has started on port: %s", appCfg.AppPort)

	<-exit
	//Graceful shutdown
	logger.Info("Shutting down gracefully...")
	//Time to shutdown gracefully
	gctx, gcancel := context.WithTimeout(context.Background(), time.Second*5)
	defer func() {
		pg.CloseConn()
		gcancel()
	}()

	if err := srv.Shutdown(gctx); err != nil {
		logger.Fatalf("server could not shutdown gracefully. %s", err.Error())
	}

	logger.Info("ok")
}
