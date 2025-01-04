package main

import (
	"fmt"
	"log"

	"github.com/pearsall-will/learn-pub-sub-starter/internal/gamelogic"
	"github.com/pearsall-will/learn-pub-sub-starter/internal/pubsub"
	"github.com/pearsall-will/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	user, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	queueName := routing.PauseKey + "." + user
	con, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	fmt.Println("Welcome, ", user)

	defer con.Close()
	newCh, que, err := pubsub.DeclareAndBind(con, "peril_direct", queueName, routing.PauseKey, pubsub.SimpleQueueTransient)
	if err != nil {
		log.Fatalf("Channel creation failed: %v", err)
	}
	defer newCh.Close()
	fmt.Printf("Queue: %+v\n", que.Name)
	gs := gamelogic.NewGameState(user)
Loop:
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
			if len(cmds) != 3 {
				fmt.Println("Usage: move location unitID")
				continue
			}
			_, err := gs.CommandMove(cmds) // We will add move handling later
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			break Loop
		default:
			fmt.Println("Unknown command")
		}
	}
}
