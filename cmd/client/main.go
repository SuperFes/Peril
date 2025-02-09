package main

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")

		gs.HandlePause(ps)
	}
}

func main() {
	fmt.Println("Starting Peril client...")

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ")
		panic(err)
	}

	defer conn.Close()

	fmt.Println("Connected to RabbitMQ")

	userName := ""

	for userName == "" {
		userName, err = gamelogic.ClientWelcome()

		if err != nil {
			fmt.Println("Please try again.")
		}
	}

	pauseRoute := routing.PauseKey + "." + userName

	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		pauseRoute,
		routing.PauseKey,
		1,
	)

	if err != nil {
		fmt.Println("Failed to declare and bind queue")
	}

	gameState := gamelogic.NewGameState(userName)

	fmt.Println("Welcome to Peril, " + userName + "!")

	for {
		pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, pauseRoute, routing.PauseKey, 1, handlerPause(gameState))

		command := gamelogic.GetInput()

		if command == nil || len(command) == 0 {
			continue
		}

		if command[0] == "help" {
			gamelogic.PrintClientHelp()
		} else if command[0] == "spawn" {
			if len(command) < 3 {
				fmt.Println("spawn <location> <type>")
			}

			err = gameState.CommandSpawn(command)
		} else if command[0] == "move" {
			if len(command) < 3 {
				fmt.Println("move <location> <unitId> (<unitId> ...)")
			}

			_, err = gameState.CommandMove(command)
		} else if command[0] == "status" {
			gameState.CommandStatus()
		} else if command[0] == "spawn" {
			if len(command) < 3 {
				fmt.Println("spawn <location> <rank>")
			}

			err = gameState.CommandSpawn(command)
		} else if command[0] == "spam" {
			if len(command) < 2 {
				fmt.Println("spam <n>")
			}

			fmt.Println("Spam no workie")
			// gameState.CommandSpam(command)
		} else if command[0] == "quit" {
			gamelogic.PrintQuit()
			break
		} else {
			fmt.Println("Unknown command")
		}

		if err != nil {
			fmt.Println(err)
		}
	}
}
