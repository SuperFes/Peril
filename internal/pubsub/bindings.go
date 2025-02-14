package pubsub

import (
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int, // an enum to represent "durable" or "transient"
	dlx bool,
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()

	var args amqp.Table

	if dlx {
		args = amqp.Table{
			"x-dead-letter-exchange": routing.ExchangePerilDeadLetter,
		}
	}

	queue, err := ch.QueueDeclare(
		queueName,
		simpleQueueType == 0, // durable
		simpleQueueType == 1, // autoDelete
		simpleQueueType == 1, // exclusive
		false,                // noWait
		args,
	)

	if err != nil {
		return nil, queue, err
	}

	err = ch.QueueBind(queue.Name, key, exchange, false, nil)

	return ch, queue, err
}
