package main

import (
	"fmt"
	"log"

	"github.com/pearsall-will/learn-pub-sub-starter/internal/gamelogic"
	"github.com/pearsall-will/learn-pub-sub-starter/internal/pubsub"
	"github.com/pearsall-will/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Handlers
func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(move)
		switch outcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap()},
			)
			if err != nil {
				fmt.Printf("Error publishing war recognition: %v", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}
		fmt.Println("Unknown outcome")
		return pubsub.NackDiscard
	}
}

func handlerWar(gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(war gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, _, _ := gs.HandleWar(war)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			return pubsub.Ack
		default:
			return pubsub.NackDiscard
		}
	}
}

// main
func main() {
	fmt.Println("Starting Peril client...")
	con, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer con.Close()

	publishCh, err := con.Channel()
	if err != nil {
		log.Fatalf("Channel failed: %v", err)
	}

	user, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	gs := gamelogic.NewGameState(user)

	err = pubsub.SubscribeJSON(
		con,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+gs.GetUsername(),
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		handlerMove(gs, publishCh),
	)
	if err != nil {
		log.Fatalf("Subscription failed: %v", err)
	}

	err = pubsub.SubscribeJSON(
		con,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.SimpleQueueDurable,
		handlerWar(gs),
	)
	if err != nil {
		log.Fatalf("Subscription failed: %v", err)
	}
	err = pubsub.SubscribeJSON(
		con,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gs.GetUsername(),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gs),
	)
	if err != nil {
		log.Fatalf("Subscription failed: %v", err)
	}

REPL:
	for {
		cmds := gamelogic.GetInput()
		if len(cmds) == 0 {
			continue
		}

		switch cmds[0] {
		case "spawn":
			if len(cmds) != 3 {
				fmt.Println("Usage: spawn location unit")
				continue
			}
			err := gs.CommandSpawn(cmds)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "move":
			mv, err := gs.CommandMove(cmds)
			if err != nil {
				fmt.Println(err)
				continue
			}
			pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+gs.GetUsername(),
				mv)

			fmt.Println("Move sent!")
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			break REPL
		default:
			fmt.Println("Unknown command")
		}
	}
}
