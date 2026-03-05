package main

import (
	"context"
	"flag"
	"media_ads/internal/config"
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

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err = service.New(redisClient, 10).Run(ctx)
	if err != nil {
		panic(err)
	}

}
