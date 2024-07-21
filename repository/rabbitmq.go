package repository

import (
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"rb-scrapper/entity"
	"time"
)

type RabbitMQ struct {
	ch       *amqp091.Channel
	exchange string
}

func NewRabbitMQ(exchange string, ch *amqp091.Channel) *RabbitMQ {
	return &RabbitMQ{exchange: exchange, ch: ch}
}

func (rbt *RabbitMQ) Store(url *entity.URL) error {
	bt, err := url.JSON()
	if err != nil {
		return err
	}
	fmt.Println("scrap." + url.Query)
	return rbt.ch.Publish(rbt.exchange, "scrap."+url.Query, false, false, amqp091.Publishing{
		ContentType: "application/json",
		Body:        bt,
		Timestamp:   time.Now(),
	})
}
