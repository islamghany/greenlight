package event

import (
	"context"
	"encoding/json"
	"fmt"
	"listener-service/mailpb"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn       *amqp.Connection
	mailClient mailpb.MailSeviceClient
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func NewConsumer(conn *amqp.Connection, mailClient mailpb.MailSeviceClient) (*Consumer, error) {

	consumer := &Consumer{
		conn:       conn,
		mailClient: mailClient,
	}

	err := consumer.setup()
	if err != nil {
		return nil, err
	}

	return consumer, nil
}

func (c *Consumer) setup() error {
	channel, err := c.conn.Channel()
	if err != nil {
		return err
	}

	return declareExchange(channel)
}

func (c *Consumer) Listen(topics []string) error {
	ch, err := c.conn.Channel()

	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := declareRandomQueue(ch)

	if err != nil {
		return err
	}

	for _, s := range topics {
		err = ch.QueueBind(
			q.Name,
			s,
			"messages_topic",
			false,
			nil,
		)
		if err != nil {
			return err
		}
	}
	messages, err := ch.Consume(q.Name, "", true, false, false, false, nil)

	if err != nil {
		return err
	}

	forerver := make(chan bool)
	go func() {
		for d := range messages {

			go c.handlePayload(d.Body)
		}
	}()

	fmt.Printf("Waiting for message [Exchange, Queue] [messages_topic, %s]\n", q.Name)
	<-forerver

	return nil
}

func (c *Consumer) handlePayload(p []byte) {
	var payload Payload

	err := json.Unmarshal(p, &payload)

	if payload.Name == "mail" && err == nil {
		err = c.sendMailViaGRPC([]byte(payload.Data))
	}
	if err != nil || payload.Name == "log" {
		err = sendLogViaGRPC([]byte(payload.Data))
		if err != nil {
			log.Println(err)
		}
	}

}

func (c *Consumer) sendMailViaGRPC(data []byte) error {

	m := mailpb.Mail{}
	err := json.Unmarshal(data, &m)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	_, err = c.mailClient.SendMail(ctx, &mailpb.MailRequest{
		MailEntry: &mailpb.Mail{
			From:         m.From,
			To:           m.To,
			Subject:      m.Subject,
			TemplateFile: m.TemplateFile,
			Data:         m.Data,
			Attachments:  []string{},
		},
	})

	if err != nil {
		return err
	}
	return nil

}
func sendLogViaGRPC(payload []byte) error {

	return nil
}
