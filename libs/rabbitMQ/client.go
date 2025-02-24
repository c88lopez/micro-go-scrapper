package rabbitMQ

import (
	"fmt"
	"log"

	"sarasa/schemas"

	"github.com/streadway/amqp"
)

type Client struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

var lastrc schemas.RabbitMQConfig

func (rabbitmq *Client) Init(rc schemas.RabbitMQConfig) error {
	log.Println("Initializing RabbitMQ client...")

	var err error

	url := fmt.Sprintf("amqp://%s:%s@%s:%d/", rc.User, rc.Password, rc.Host, rc.Port)

	rabbitmq.Connection, err = amqp.Dial(url)
	if err != nil {
		return err
	}

	rabbitmq.Channel, err = rabbitmq.Connection.Channel()
	if err != nil {
		return err
	}

	lastrc = rc

	return nil
}

func (rabbitmq *Client) Restart() error {
	rabbitmq.Close()
	return rabbitmq.Init(lastrc)
}

func (rabbitmq Client) Close() error {
	rabbitmq.Channel.Close()
	rabbitmq.Channel = nil

	rabbitmq.Connection.Close()
	rabbitmq.Connection = nil

	return nil
}

func (rabbitmq Client) String() string {
	return "RabbitMQ"
}
