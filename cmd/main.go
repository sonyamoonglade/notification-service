package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sonyamoonglade/notification-service/config"
	"github.com/sonyamoonglade/notification-service/internal/app_middlewares"
	"github.com/sonyamoonglade/notification-service/internal/events"
	"github.com/sonyamoonglade/notification-service/internal/storage"
	"github.com/sonyamoonglade/notification-service/internal/subscription"
	"github.com/sonyamoonglade/notification-service/pkg/bot"
	"github.com/sonyamoonglade/notification-service/pkg/formatter"
	"github.com/sonyamoonglade/notification-service/pkg/logging"
	"github.com/sonyamoonglade/notification-service/pkg/postgres"
	"github.com/sonyamoonglade/notification-service/pkg/server"
	"github.com/sonyamoonglade/notification-service/pkg/telegram"
	"github.com/sonyamoonglade/notification-service/pkg/template"
)

func main() {

	log.Println("booting an application")

	logsPath, debug, strictMode := parseFlags()

	logger, err := logging.WithConfig(&logging.Config{
		Strict:   strictMode,
		LogsPath: logsPath,
		Debug:    debug,
		Encoding: logging.JSON,
	})
	if err != nil {
		log.Fatalf("could not get logger. %s", err.Error())
	}

	//Make sure to load env variables on local development
	if err := godotenv.Load(".env.local"); err != nil {
		logger.Warnf("could not load .env.local %s", err.Error())
	}

	appCfg, err := config.GetAppConfig()
	if err != nil {
		logger.Fatalf("could not get app config. %s", err.Error())
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

	pgStorage := storage.NewPostgresStorage(logger, pg.Pool)

	eventsService := events.NewEventsService(logger, pgStorage, templateProvider)

	mw := app_middlewares.New(logger, eventsService)

	subscriptionService := subscription.NewSubscriptionService(logger, pgStorage)
	subscriptionTransport := subscription.NewSubscriptionTransport(logger,
		subscriptionService,
		mw.DoesExist,
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
	logger.Info("bot is listening to updates and ready to notify")

	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("could not start server. %s", err.Error())
		}

	}()
	logger.Infof("server has started on port: %s", appCfg.AppPort)

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	<-exit
	//Graceful shutdown
	logger.Info("Shutting down gracefully...")

	//Timeout for shutdown
	gctx, gcancel := context.WithTimeout(context.Background(), time.Second*5)
	defer gcancel()

	defer func() {
		logger.Info("before closing postgres")
		pg.CloseConn()
		logger.Info("closing postgres connection...")

		appBot.ClosePoll()
		logger.Info("closing bot poll...")
	}()

	if err := srv.Shutdown(gctx); err != nil {
		logger.Fatalf("server could not shutdown gracefully. %s", err.Error())
	}
	logger.Info("server has shutdown")

}

func parseFlags() (string, bool, bool) {

	logsPath := flag.String("logs-path", "", "defines path to logging file")
	debug := flag.Bool("debug", true, "defines debug mode")
	strictMode := flag.Bool("strict", false, "defines strictness of the logs")

	flag.Parse()

	return *logsPath, *debug, *strictMode
}
