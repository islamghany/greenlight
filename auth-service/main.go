package main

import (
	"auth-service/api"
	"auth-service/token"
	"auth-service/utils"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	"auth-service/db/cache"
	sqlc "auth-service/db/sqlc"

	"github.com/go-redis/redis/v8"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {

	//  load the env variables
	config, err := utils.LoadConfig(".")

	if err != nil {
		log.Fatal(err)
	}
	flag.IntVar(&config.DB_MAX_OPEN_CONNECTION, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&config.DB_MAX_IDLE_CONNECTION, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&config.DB_MAX_IDLE_TIME, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.Parse()

	// initlize the user struct validator
	v, err := utils.NewUserValidator()

	if err != nil {
		log.Fatal(err)
	}

	//  connect to the db
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

	store := sqlc.New(db)
	// connect to redis caching
	rdb, err := utils.Connect("redis", 10, 1*time.Second, func() (*redis.Client, error) {
		return openRedis(config.REDIS_HOST, config.REDIS_PORT)
	})

	if err != nil {
		log.Fatal(err)
	}
	redisCache := cache.NewCache(rdb)
	defer rdb.Close()

	maker, err := token.NewPasetoMaker(config.TOKEN_SYMMETRIC_KEY[:32])

	if err != nil {
		log.Fatal(err)
	}
	server := api.NewServer(store, redisCache, &config, v, maker)

	log.Printf("Connected to server on port %d \n", config.PORT)
	go server.Start(config.PORT)

	log.Println("GRPC IS UP")
	err = server.OpenGRPC(50051)
	if err != nil {
		log.Fatal(err)
	}

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

func openRedis(host, port string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Password: "",
		Addr:     fmt.Sprint(host, ":", port),
		DB:       0,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := rdb.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}
	return rdb, nil
}
