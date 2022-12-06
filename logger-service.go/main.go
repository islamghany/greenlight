package main

import (
	"context"
	"fmt"
	"log"
	"logger-service/api"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	mongoURL = "mongodb://mongo:27017"
	gRpcPort = 50051
)

func main() {

	// connect to redis caching
	db, err := Connect("Mongo", 10, 1*time.Second, func() (*mongo.Client, error) {
		return connectToMongo()
	})
	if err != nil {
		log.Fatal(err)
	}

	server := api.NewServer(db)

	fmt.Println("Running the Logger service grpc")
	err = server.OpenGRPC(gRpcPort)
	if err != nil {
		log.Fatal(err)
	}

}

func connectToMongo() (*mongo.Client, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURL)

	c, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	if err := c.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return c, nil
}
func Connect[T any](connectName string, counts int64, backOff time.Duration, fn func() (*T, error)) (*T, error) {
	var connection *T

	for {
		c, err := fn()
		if err == nil {
			log.Println("connected to: ", connectName)
			connection = c
			break
		}

		log.Printf("%s not yet read", connectName)
		counts--
		if counts == 0 {
			return nil, fmt.Errorf("can not connect to the %s", connectName)
		}
		backOff = backOff + (time.Second * 2)

		log.Println("Backing off.....")
		time.Sleep(backOff)
		continue

	}
	return connection, nil
}
