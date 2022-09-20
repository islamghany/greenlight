package event

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Emitter struct {
	connetion *amqp.Connection
}

func NewEventEmitter(conn *amqp.Connection) (*Emitter, error) {
	e := Emitter{
		connetion: conn,
	}
	err := e.setup()

	if err != nil {
		return nil, err
	}

	return &e, nil
}
func (e *Emitter) setup() error {
	ch, err := e.connetion.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	return declareExchange(ch)
}

func (e *Emitter) Push(event string, severity string) error {
	ch, err := e.connetion.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	log.Println("Pushing to channel")
	err = ch.Publish(
		"logs_topic",
		severity,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(event),
		},
	)
	if err != nil {
		return err
	}
	return nil
}
