package event

import (
	"encoding/json"
	"fmt"
	"log"

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

			go handlePayload(d.Body)
		}
	}()

	fmt.Printf("Waiting for message [Exchange, Queue] [messages_topic, %s]\n", q.Name)
	<-forerver

	return nil
}

func handlePayload(p []byte) {

	var payload Payload

	err := json.Unmarshal(p, &payload)

	if payload.Name == "mail" && err == nil {
		err = sendMailViaGRPC([]byte(payload.Data))
	}
	if err != nil || payload.Name == "log" {
		err = sendLogViaGRPC([]byte(payload.Data))
		if err != nil {
			log.Println(err)
		}
	}

}

type MailPayload struct {
	From         string `json:"from"`
	To           string `json:"to"`
	Subject      string `json:"subject"`
	Message      string `json:"message"`
	TemplateFile string `json:"templateFile"`
}

func sendMailViaGRPC(payload []byte) error {

	log.Println(string(payload))
	return nil
	// conn, err := grpc.Dial("mail-service:50051", grpc.WithInsecure())

	// if err != nil {
	// 	return err
	// }

	// defer conn.Close()

	// c := mailpb.NewMailSeviceClient(conn)

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)

	// defer cancel()
	// dest := &MailPayload{}
	// err = json.Unmarshal(payload, dest)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }

	// _, err = c.SendMail(ctx, &mailpb.MailRequest{
	// 	MailEntry: &mailpb.Mail{
	// 		From:         dest.From,
	// 		To:           dest.To,
	// 		Subject:      dest.Subject,
	// 		TemplateFile: dest.TemplateFile,
	// 		Message:      dest.Message,
	// 		Attachments:  []string{},
	// 	},
	// })

	// if err != nil {
	// 	return err
	// }
	// return nil
}
func sendLogViaGRPC(payload []byte) error {

	return nil
}
