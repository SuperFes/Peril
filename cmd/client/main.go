package main

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
	"strconv"
	"time"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")

		gs.HandlePause(ps)

		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, conn *amqp.Connection) func(move gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")

		outcome := gs.HandleMove(move)

		if outcome == gamelogic.MoveOutcomeSafe {
			return pubsub.Ack
		} else if outcome == gamelogic.MoveOutcomeMakeWar {
			fmt.Println("War!")

			ch, _, err := pubsub.DeclareAndBind(
				conn,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix,
				routing.WarRecognitionsPrefix+".*",
				0,
				true,
			)

			if err != nil {
				fmt.Println(err)
			}

			err = pubsub.PublishJSON(ch, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix+"."+gs.GetUsername(), gamelogic.RecognitionOfWar{
				Attacker: move.Player,
				Defender: gs.Player,
			})

			if err != nil {
				fmt.Println(err)

				return pubsub.NackRequeue
			}

			return pubsub.Ack
		}

		return pubsub.NackDiscard
	}
}

func handlerWar(gs *gamelogic.GameState, conn *amqp.Connection) func(decl gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(decl gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")

		draw, win, lose := gs.HandleWar(decl)

		ch, _, err := pubsub.DeclareAndBind(
			conn,
			routing.ExchangePerilTopic,
			routing.GameLogSlug,
			routing.GameLogSlug+"."+gs.GetUsername(),
			0,
			false,
		)

		if err != nil {
			fmt.Println(err)
		}

		msg := fmt.Sprintf("A war between %s and %s resulted in a draw", win, lose)

		if draw == gamelogic.WarOutcomeYouWon || draw == gamelogic.WarOutcomeOpponentWon {
			msg = fmt.Sprintf("%s won a war against %s", win, lose)
		}

		err = pubsub.PublishGob(ch, routing.ExchangePerilTopic, routing.GameLogSlug+"."+gs.GetUsername(), routing.GameLog{
			CurrentTime: time.Now(),
			Message:     msg,
			Username:    decl.Attacker.Username,
		})

		if err != nil {
			return pubsub.NackRequeue
		}

		if draw == gamelogic.WarOutcomeNotInvolved {
			return pubsub.NackRequeue
		} else if draw == gamelogic.WarOutcomeNoUnits {
			return pubsub.NackDiscard
		} else if draw == gamelogic.WarOutcomeOpponentWon ||
			draw == gamelogic.WarOutcomeYouWon ||
			draw == gamelogic.WarOutcomeDraw {
			return pubsub.Ack
		}

		fmt.Println("Error, unknown thing happened in war")

		return pubsub.NackDiscard
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
	movesRoute := routing.ArmyMovesPrefix + "." + userName

	ch, _, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		pauseRoute,
		routing.PauseKey,
		1,
		true,
	)

	if err != nil {
		fmt.Println("Failed to declare and bind queue")
		panic(err)
	}

	var move gamelogic.ArmyMove

	gameState := gamelogic.NewGameState(userName)

	fmt.Println("Welcome to Peril, " + userName + "!")

	pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, pauseRoute, routing.PauseKey, 1, handlerPause(gameState))
	pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, movesRoute, routing.ArmyMovesPrefix+".*", 1, handlerMove(gameState, conn))
	pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, routing.WarRecognitionsPrefix+".*", 0, handlerWar(gameState, conn))

	for {
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

			move, err = gameState.CommandMove(command)

			if err != nil {
				fmt.Println(err)
			}

			err = pubsub.PublishJSON(ch, routing.ExchangePerilTopic, movesRoute, move)
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

			count, err := strconv.Atoi(command[1])

			if err != nil {
				fmt.Println("Please enter a valid number")
				continue
			}

			userName := gameState.GetUsername()

			fmt.Println("Spamming...")

			for i := 0; i < count; i++ {
				err = pubsub.PublishGob(ch, routing.ExchangePerilTopic, routing.GameLogSlug+"."+userName, routing.GameLog{
					CurrentTime: time.Now(),
					Message:     gamelogic.GetMaliciousLog(),
					Username:    userName,
				})

				if err != nil {
					fmt.Println(err)
				}
			}

			fmt.Println("Spammed!")
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
