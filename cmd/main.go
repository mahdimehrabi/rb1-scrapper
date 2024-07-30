package main

import (
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"log"
	"log/slog"
	"net/http"
	"os"
	"rb-scrapper/entity"
	"rb-scrapper/infrastructure/godotenv"
	"rb-scrapper/repository"
	"rb-scrapper/utils"
	"strconv"
	"time"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdin, nil))
	iuc := make(chan *entity.URL, 10)

	env := godotenv.NewEnv()
	env.Load()
	fmt.Println(env)
	ampqConn, err := amqp091.Dial(env.AMQP)
	FatalOnError(err)
	defer ampqConn.Close()
	ch, err := ampqConn.Channel()
	FatalOnError(err)
	err = ch.ExchangeDeclare(env.ImageExchange, "topic", true, false, false, false, nil)
	FatalOnError(err)

	rbt := repository.NewRabbitMQ(env.ImageExchange, ch)

	go initHTTPServer(env, logger, iuc)

	for i := range iuc {
		err = rbt.Store(i)
		if err != nil {
			logger.Error("error producing", err.Error())
		}
		logger.Info("sent", i.URL)
	}
}

func initHTTPServer(env *godotenv.Env, logger *slog.Logger, iuc chan *entity.URL) *http.Server {
	mux := http.NewServeMux()
	httpServer := &http.Server{
		Addr:        env.Host,
		Handler:     mux,
		ReadTimeout: 3 * time.Second,
	}
	mux.HandleFunc("/{count}", func(w http.ResponseWriter, r *http.Request) {
		countStr := r.PathValue("count")
		count, err := strconv.Atoi(countStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		drs := utils.NewDownloadResizer(count, logger, env.ScrapTopics)
		go drs.Download(iuc)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Error("failed to run app", "err", err)
			panic(err)
		}
	}()
	fmt.Printf("\nrunning on %s", env.Host)
	return httpServer
}
func FatalOnError(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}
