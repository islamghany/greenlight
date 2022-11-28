package main

import (
	"auth-service/utils"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	_ "github.com/go-redis/redis/v8"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	_ "github.com/rabbitmq/amqp091-go"
)

func main() {

	// 1- load the env variables
	config, err := utils.LoadConfig(".")

	if err != nil {
		log.Fatal(err)
	}
	flag.IntVar(&config.DB_MAX_OPEN_CONNECTION, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&config.DB_MAX_IDLE_CONNECTION, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&config.DB_MAX_IDLE_TIME, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.Parse()

	// 2- connect to the db
	db, err := utils.Connect("postgres", 10, 1*time.Second, func() (*sql.DB, error) {
		return openDB(&config)
	})

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	// 3- run the database migrations

	err = runMigrate(config.DSN, config.MIGRATION_URL)
	if err != nil {
		log.Fatal(err)
	}
	// connect to the grpc

	// start to listen on a port
	// server := api.NewServer(db)
	// log.Printf("Connected to server on port %s \n", conf.port)
}
func openDB(config *utils.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.DSN)

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(config.DB_MAX_OPEN_CONNECTION)
	db.SetMaxIdleConns(config.DB_MAX_IDLE_CONNECTION)
	// to a time.Duration type.
	duration, err := time.ParseDuration(config.DB_MAX_IDLE_TIME)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func runMigrate(dsn, migrationPath string) error {

	migration, err := migrate.New("file://db/migrations", dsn)

	if err != nil {
		return err
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrate up: %w", err)
	}

	log.Println("Successfully migrated db")
	return nil
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
