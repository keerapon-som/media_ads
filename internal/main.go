package main

import (
	"context"
	"flag"
	"media_ads/internal/config"
	"media_ads/internal/domain"
	"media_ads/internal/repository"
	"media_ads/internal/service"
	"media_ads/packages"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
)

var (
	log = logrus.New()

	debug = flag.Bool("debug", false, "-debug")
)

func initLogConfig(debug bool) {
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	log.SetFormatter(customFormatter)

	if debug {
		log.Level = logrus.DebugLevel
		log.Printf("Log level : %s", log.Level.String())
	}
}

func main() {

	config := config.GetConfig()

	redisClient, err := packages.NewRedisClient(
		config.ServiceConfig.Redis.Addr,
		config.ServiceConfig.Redis.Password,
		config.ServiceConfig.Redis.DB,
	)
	if err != nil {
		log.WithError(err).Fatal("failed to initialize redis client")
	}
	if err != nil {
		panic(err)
	}
	initLogConfig(*debug)

	objectFileTransfer := packages.NewObjectFileTransferLocal("root-object-download")

	pg := packages.NewPostgresql(
		config.ServiceConfig.PostgresConfig.Host,
		config.ServiceConfig.PostgresConfig.Port,
		config.ServiceConfig.PostgresConfig.User,
		config.ServiceConfig.PostgresConfig.Password,
		config.ServiceConfig.PostgresConfig.DBName,
		config.ServiceConfig.PostgresConfig.SSLMode,
	)

	db, err := pg.Connect()
	if err != nil {
		log.WithError(err).Fatal("failed to connect to postgres")
	}
	defer packages.Disconnect(db)

	mediaProviderRepo := repository.NewMediaArchiveRepo(db)

	mediaPublisher := domain.NewMediaPublisher()
	mediaProvider := domain.NewMediaArchive(objectFileTransfer, mediaProviderRepo)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err = service.New(redisClient, mediaPublisher, mediaProvider, 10).Run(ctx)
	if err != nil {
		panic(err)
	}

}
