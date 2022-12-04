package main

import (
	"errors"
	"fmt"
	"listener-service/event"
	"listener-service/mailpb"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

func main() {

	// connected to rabbitmq (broker)

	rabbitConn, err := connect(10, 1*time.Second)

	if err != nil {
		log.Fatal(err)
	}
	defer rabbitConn.Close()

	// connect to the auth servie via grpc
	conn, err := grpc.Dial("mail-service:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	m := mailpb.NewMailSeviceClient(conn)
	// start listening for message
	consumer, err := event.NewConsumer(rabbitConn, m)
	if err != nil {
		log.Fatal(err)
	}

	// watch the queue and conume events
	err = consumer.Listen([]string{"mail"})
	if err != nil {
		log.Println(err)
	}
}

func connect(counts int64, backOff time.Duration) (*amqp.Connection, error) {
	var connection *amqp.Connection

	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err == nil {
			log.Println("connected to RabbitMQ")
			connection = c
			break
		}

		fmt.Println("RabbitMQ not yet read")
		counts--
		if counts == 0 {
			return nil, errors.New("Can not connect to the RabbitMQ")
		}
		backOff = backOff + (time.Second * 2)

		fmt.Println("Backing off.....")
		time.Sleep(backOff)
		continue

	}
	return connection, nil
}
