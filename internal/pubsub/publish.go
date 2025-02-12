package pubsub

import (
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
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
	handler func(T) AckType,
) error {
	go func() {
		for {
			ch, _, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)

			if err != nil {
				panic(err)
			}

			msgs, err := ch.Consume(queueName, "", true, false, false, false, nil)

			if err != nil {
				panic(err)
			}

			for d := range msgs {
				var val T
				err := json.Unmarshal(d.Body, &val)

				if err != nil {
					fmt.Println("Unable to unmarshal JSON")
					fmt.Println(err)
					continue
				}

				dakTakLakPak := handler(val)

				if dakTakLakPak == Ack {
					fmt.Println("Ack")

					d.Ack(false)
				} else if dakTakLakPak == NackRequeue {
					fmt.Println("NackRequeue")

					d.Nack(false, true)
				} else if dakTakLakPak == NackDiscard {
					fmt.Println("NackDiscard")

					d.Nack(false, false)
				}
			}

			ch.Close()
		}
	}()

	return nil
}
