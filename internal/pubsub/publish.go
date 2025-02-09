package pubsub

import (
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	data, err := json.Marshal(val)

	if err != nil {
		fmt.Println("Unable to marshal JSON")
		fmt.Println(err)

		return err
	}

	//ctx := context.Background()

	mandatory := false
	immediate := false

	msg := amqp.Publishing{
		ContentType: "application/data",
		Body:        data,
	}

	return ch.Publish(exchange, key, mandatory, immediate, msg)
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
	handler func(T),
) error {
	ch, _, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)

	if err != nil {
		return err
	}

	msgs, err := ch.Consume(queueName, "", true, false, false, false, nil)

	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			var val T
			err := json.Unmarshal(d.Body, &val)

			if err != nil {
				fmt.Println("Unable to unmarshal JSON")
				fmt.Println(err)
				continue
			}

			d.Ack(false)

			handler(val)
		}

		defer ch.Close()
	}()

	return nil
}
