package main

import (
	"fmt"
	"os"
	"os/signal"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	fmt.Println("Starting Peril server...")
	con, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		fmt.Println("Connection failed!!!")
		fmt.Println(err)
	} else {
		defer con.Close()
		fmt.Println("Connection successful...")
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		<-sigChan
		fmt.Println("Closing Terminal...")
	}
}
