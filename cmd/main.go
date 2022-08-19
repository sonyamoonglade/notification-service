package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sonyamoonglade/delivery-service/pkg/logging"
	"github.com/sonyamoonglade/notification-service/config"
	"github.com/sonyamoonglade/notification-service/internal/events"
	"github.com/sonyamoonglade/notification-service/internal/events/middleware"
	"github.com/sonyamoonglade/notification-service/internal/subscription"
	"github.com/sonyamoonglade/notification-service/pkg/bot"
	"github.com/sonyamoonglade/notification-service/pkg/formatter"
	"github.com/sonyamoonglade/notification-service/pkg/postgres"
	"github.com/sonyamoonglade/notification-service/pkg/server"
	"github.com/sonyamoonglade/notification-service/pkg/telegram"
	"github.com/sonyamoonglade/notification-service/pkg/template"
	"go.uber.org/zap"
)

func main() {

	log.Println("booting an application")

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	//Todo: get dev mode from env
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

	appBot, err := bot.NewBot(appCfg.BotToken, logger)
	if err != nil {
		logger.Fatalf("could not create bot instance. %s", err.Error())
	}

	appFmt := formatter.NewFormatter()
	templateProvider := template.NewTemplateProvider()

	//Read templates.json
	if err = templateProvider.ReadTemplates(); err != nil {
		logger.Fatalf("could not read templates. %s", err.Error())
	}

	eventsStorage := events.NewEventStorage(logger, pg.Pool)
	eventsService := events.NewEventsService(logger, eventsStorage, templateProvider)
	eventsMiddleware := middleware.NewEventsMiddlewares(logger, eventsService)

	subscriptionStorage := subscription.NewSubscriptionStorage(logger, pg.Pool)
	subscriptionService := subscription.NewSubscriptionService(logger, subscriptionStorage)

	subscriptionTransport := subscription.NewSubscriptionTransport(logger,
		subscriptionService,
		eventsMiddleware,
		eventsService,
		templateProvider,
		appFmt,
		appBot)

	telegramListener := telegram.NewTelegramListener(logger, appBot, subscriptionService)

	subscriptionTransport.InitRoutes(router)
	logger.Info("initialized routes")

	//Read events.json
	if err = eventsService.ReadEvents(ctx); err != nil {
		logger.Fatalf("could not load base events. %s", err.Error())
	}

	go telegramListener.ListenForUpdates()
	logger.Info("notification bot is listening to updates and ready to notify")

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
