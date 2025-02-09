package pubsub

import (
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	json, err := json.Marshal(val)

	if err != nil {
		fmt.Println("Unable to marshal JSON")
		fmt.Println(err)

		return err
	}

	//ctx := context.Background()

	mandatory := false
	immediate := false

	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        json,
	}

	return ch.Publish(exchange, key, mandatory, immediate, msg)
}
