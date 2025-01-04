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

	fmt.Println("Starting Peril server...")
	gamelogic.PrintServerHelp()

	con, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer con.Close()

	// Setup Channel
	fmt.Println("Connection successful...")
	fmt.Println("Creating Channel...")
	ch, err := con.Channel()
	if err != nil {
		log.Fatalf("Channel failed: %v", err)
	}
	// New Queue
	newCh, que, err := pubsub.DeclareAndBind(
		con,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.SimpleQueueDurable)
	if err != nil {
		log.Fatalf("Channel creation failed: %v", err)
	}
	defer newCh.Close()
	fmt.Printf("Queue %+v created.\n", que.Name)

	// Server Loop
Loop:
	for {
		words := gamelogic.GetInput()
		switch words[0] {
		case "pause":
			fmt.Println("Sending Pause...")
			err = pubsub.PublishJSON(
				ch,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: true},
			)
			if err != nil {
				log.Printf("Publish failed: %v", err)
			}
		case "resume":
			fmt.Println("Sending Resume...")
			err = pubsub.PublishJSON(
				ch,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: false},
			)
			if err != nil {
				log.Printf("Publish failed: %v", err)
			}
		case "quit":
			fmt.Println("Closing Terminal...")
			break Loop
		default:
			fmt.Println("Unknown command")
		}
	}
}
