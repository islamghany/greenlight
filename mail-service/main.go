package main

import (
	"log"
	"mailer-service/api"
)

func main() {
	server, err := api.NewServer()
	if err != nil {
		log.Fatal(err)
	}
	server.Start(8080)
	if err != nil {
		log.Fatal(err)
	}
}
