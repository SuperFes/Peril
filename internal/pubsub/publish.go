package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)
	err := enc.Encode(val)

	if err != nil {
		log.Fatal("encode:", err)
	}

	mandatory := false
	immediate := false

	msg := amqp.Publishing{
		ContentType: "application/gob",
		Body:        buf.Bytes(),
	}

	return ch.Publish(exchange, key, mandatory, immediate, msg)
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
	handler func(T) AckType,
	unmarshaller func([]byte) (T, error),
) error {
	go func() {
		for {
			ch, _, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType, false)

			if err != nil {
				panic(err)
			}

			ch.Qos(10, 0, false)

			msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)

			if err != nil {
				panic(err)
			}

			for d := range msgs {
				var val T
				val, err = unmarshaller(d.Body)

				//if err != nil {
				//	fmt.Println("Unable to unmarshal Gob")
				//	fmt.Println(err)
				//	d.Nack(false, true)
				//	continue
				//}

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

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	data, err := json.Marshal(val)

	if err != nil {
		fmt.Println("Unable to marshal JSON")
		fmt.Println(err)

		return err
	}

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
			ch, _, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType, true)

			if err != nil {
				panic(err)
			}

			ch.Qos(10, 0, false)

			msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)

			if err != nil {
				panic(err)
			}

			for d := range msgs {
				var val T
				_ = json.Unmarshal(d.Body, &val)

				//if err != nil {
				//	fmt.Println("Unable to unmarshal JSON")
				//	fmt.Println(err)
				//	d.Nack(false, true)
				//	continue
				//}

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
