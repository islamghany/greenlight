package event

import (
	"context"
	"encoding/json"
	"time"

	"auth-service/mailpb"

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
	// create a channel
	ch, err := e.connetion.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// here i declare a queue, if the queue doesn't exist it will declare it.
	return declareExchange(ch)
}

func (e *Emitter) Push(event string, severity string) error {
	ch, err := e.connetion.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = ch.PublishWithContext(
		ctx,
		"messages_topic",
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

func (e *Emitter) SendToMailService(m *mailpb.Mail) error {

	mailJSON, err := json.Marshal(m)

	if err != nil {
		return err
	}

	payload := Payload{
		Name: "mail",
		Data: string(mailJSON),
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	err = e.Push(string(payloadJSON), payload.Name)
	if err != nil {
		return err
	}

	return nil
}
