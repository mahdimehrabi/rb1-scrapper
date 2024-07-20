package main

import (
	"github.com/rabbitmq/amqp091-go"
	"log"
	"log/slog"
	"os"
	"rb-scrapper/entity"
	"rb-scrapper/infrastructure/godotenv"
	"rb-scrapper/repository"
	"rb-scrapper/utils"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdin, nil))
	drs := utils.NewDownloadResizer(21, logger, []string{"cat"})
	iuc := make(chan *entity.URL, 10)

	env := godotenv.NewEnv()
	env.Load()
	ampqConn, err := amqp091.Dial(env.AMQP)
	FatalOnError(err)
	defer ampqConn.Close()
	ch, err := ampqConn.Channel()
	FatalOnError(err)
	err = ch.ExchangeDeclare(env.ImageExchange, "topic", true, false, false, false, nil)
	FatalOnError(err)

	rbt := repository.NewRabbitMQ(env.ImageExchange, ch)
	go drs.Download(iuc)

	for i := range iuc {
		err = rbt.Store(i)
		if err != nil {
			logger.Error("error producing", err.Error())
		}
		logger.Info("sent", i.URL)
	}
}

func FatalOnError(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}
