package event

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"listener-service/mailpb"
	"log"
	"net/http"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
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
			go handlePayload(d.Body)
		}
	}()

	fmt.Printf("Waiting for message [EXchange, Queue] [logs_topic, %s]\n", q.Name)
	<-forerver

	return nil
}

func handlePayload(p []byte) {
	var err error

	err = sendMail(p)
	if err != nil {
		log.Println(err)
	}
	// switch p.Name {
	// case "mail", "event":
	// 	err = sendMailViaGRPC(p)
	// 	if err != nil {
	// 		log.Println(err)
	// 	}
	// default:
	// 	err = sendMailViaGRPC(p)
	// 	if err != nil {
	// 		log.Println(err)
	// 	}
	// }
}
func logEvent(entry Payload) error {
	return nil
}

type MailPayload struct {
	From         string `json:"from"`
	To           string `json:"to"`
	Subject      string `json:"subject"`
	Message      string `json:"message"`
	TemplateFile string `json:"templateFile"`
}

func sendMailViaGRPC(payload []byte) error {
	conn, err := grpc.Dial("mail-service:50051", grpc.WithInsecure())

	if err != nil {
		return err
	}

	defer conn.Close()

	c := mailpb.NewMailSeviceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)

	defer cancel()
	dest := &MailPayload{}
	err = json.Unmarshal(payload, dest)
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = c.SendMail(ctx, &mailpb.MailRequest{
		MailEntry: &mailpb.Mail{
			From:         dest.From,
			To:           dest.To,
			Subject:      dest.Subject,
			TemplateFile: dest.TemplateFile,
			Message:      dest.Message,
			Attachments:  []string{},
		},
	})

	if err != nil {
		return err
	}
	return nil
}
func sendMail(entry []byte) error {

	mailServiceURL := "http://mail-service/send"

	// post to mail service
	request, err := http.NewRequest("POST", mailServiceURL, bytes.NewBuffer(entry))
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
