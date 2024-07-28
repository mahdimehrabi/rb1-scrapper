package godotenv

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strings"
)

type Env struct {
	AMQP          string
	ImageExchange string
	ScrapTopics   []string
	Host          string
}

func NewEnv() *Env {
	return &Env{}
}

func (e *Env) Load() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}

	e.AMQP = os.Getenv("AMQP")
	e.ImageExchange = os.Getenv("IMAGE_EXCHANGE")
	e.Host = os.Getenv("HOST")
	scrapTopics := os.Getenv("SCRAP_TOPICS")
	for _, topic := range strings.Split(scrapTopics, ",") {
		e.ScrapTopics = append(e.ScrapTopics, topic)
	}
}
