package utils

import (
	"fmt"
	"log"
	"time"
)

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
