package pubsub

import amqp "github.com/rabbitmq/amqp091-go"

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()

	args := amqp.Table{}

	args["x-dead-letter-exchange"] = "peril_dlx"
	//args["x-dead-letter-routing-key"] = key

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
