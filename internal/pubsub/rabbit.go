package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleSimpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {
	ch, que, err := DeclareAndBind(conn, exchange, queueName, key, simpleSimpleQueueType)
	if err != nil {
		return err
	}
	nCh, err := ch.Consume(que.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for msg := range nCh {
			var val T
			err := json.Unmarshal(msg.Body, &val)
			if err != nil {
				log.Printf("Error unmarshalling: %v", err)
			}
			acktype := handler(val)
			switch acktype {
			case Ack:
				msg.Ack(false)
				fmt.Println("Roger Roger")
			case NackRequeue:
				msg.Nack(false, true)
				fmt.Println("No Roger Requeue")
			case NackDiscard:
				msg.Nack(false, false)
				fmt.Println("No Roger Discard")
			}
		}
	}()
	return nil
}

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	body, err := json.Marshal(val)
	if err != nil {
		return err
	}
	err = ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{ContentType: "application/json", Body: body},
	)
	return err
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueuetype SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	queue, err := ch.QueueDeclare(
		queueName,
		simpleQueuetype == SimpleQueueDurable,
		simpleQueuetype == SimpleQueueTransient,
		simpleQueuetype == SimpleQueueTransient,
		false,
		amqp.Table{"x-dead-letter-exchange": "peril_dlx"},
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	err = ch.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	return ch, queue, nil
}
