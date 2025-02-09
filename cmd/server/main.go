package main

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ")
		panic(err)
	}

	defer conn.Close()

	fmt.Println("Connected to RabbitMQ")

	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		0,
	)

	// Create a channel
	ch, err := conn.Channel()

	gamelogic.PrintServerHelp()

	for {
		command := gamelogic.GetInput()

		if command[0] == "help" {
			gamelogic.PrintServerHelp()
		} else if command[0] == "pause" {
			fmt.Println("Pausing...")
			pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})
		} else if command[0] == "unpause" || command[0] == "resume" {
			fmt.Println("Unpausing...")
			pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})
		} else if command[0] == "exit" || command[0] == "quit" {
			break
		} else {
			fmt.Println("Unknown command")
		}
	}

	fmt.Println("Shutting down Peril server...")
}
