package main

import (
	"fmt"
	"log"
	"mailer-service/api"
	"mailer-service/mailer"
	"os"
	"strconv"
)

const webPort = 80
const gRPCPort = 50051

func main() {

	port, err := strconv.Atoi(os.Getenv("MAIL_PORT"))
	if err != nil {
		log.Fatal(err)
	}

	m := mailer.Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Host:        os.Getenv("MAIL_HOST"),
		Port:        port,
		Username:    os.Getenv("MAIL_USERNAME"),
		Password:    os.Getenv("MAIL_PASSWORD"),
		Encryption:  os.Getenv("ENCRYPTION"),
		FromAddress: os.Getenv("FROM_ADDRESS"),
		FromName:    os.Getenv("FORM_NAME"),
	}

	server, err := api.NewServer(m)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Running the mail service")

	go server.Start(webPort)

	fmt.Println("Running the mail service grpc")
	err = server.OpenGRPC(gRPCPort)
	if err != nil {
		log.Fatal(err)
	}
}

////---	DEVELOPMENT
// m := mailer.Mail{
// 	Domain:      "localhost",
// 	Host:        "localhost",
// 	Port:        1025,
// 	Username:    os.Getenv("MAIL_USERNAME"),
// 	Password:    os.Getenv("MAIL_PASSWORD"),
// 	Encryption:  "none",
// 	FromAddress: "john.smith@example.com",
// 	FromName:    "john Smith",
// }
