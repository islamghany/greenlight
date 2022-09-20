package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func NewConsumer(conn *amqp.Connection) (*Consumer, error) {

	consumer := &Consumer{
		conn: conn,
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
			"logs_topic",
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
			var p Payload
			_ = json.Unmarshal(d.Body, &p)

			go handlePayload(p)
		}
	}()

	fmt.Printf("Waiting for message [EXchange, Queue] [logs_topic, %s]\n", q.Name)
	<-forerver

	return nil
}

func handlePayload(p Payload) {
	var err error
	switch p.Name {
	case "mail", "event":
		err = sendMail(p)
		if err != nil {
			log.Println(err)
		}
	default:
		err = logEvent(p)

	}
}
func logEvent(entry Payload) error {
	return nil
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func sendMail(entry Payload) error {

	jsonData, err := json.MarshalIndent(entry, "", "\t")
	if err != nil {
		return err
	}

	// call the mail service
	mailServiceURL := "http://mail-service/send"

	// post to mail service
	request, err := http.NewRequest("POST", mailServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// make sure we get back the right status code
	if response.StatusCode != http.StatusCreated {
		return err
	}

	return nil
}
