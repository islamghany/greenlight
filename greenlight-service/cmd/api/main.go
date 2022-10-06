package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"islamghany.greenlight/internals/data"
	"islamghany.greenlight/internals/jsonlog"
)

// version of the application
const version = "1.0.0"

// Create a buildTime variable to hold the executable binary build time. Note that this
// must be a string type, as the -X linker flag will only work with string variables.
var buildTime string

// config
type config struct {
	port int    // port number e.g. 8080
	env  string // development ot prodction
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}

	redis struct {
		host     string
		port     string
		username string
		password string
	}
	vars struct {
		dbDSN                     string
		emailPassword             string
		greenlightUserTokenCookie string
		greenlightUserIDCookie    string
		redisHost                 string
		redisPort                 string
		redisPassword             string
		clientUrl                 string
		greenlightEmail           string
		cldName                   string
		cldSecret                 string
		cldAPIKey                 string
	}
}

// app struct to hold the http handlers, helpers and middleware
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	wg     sync.WaitGroup
	amqp   *amqp.Connection
	cld    *cloudinary.Cloudinary
}

func main() {
	var conf config

	loadEnvVars(&conf)

	flag.IntVar(&conf.port, "port", 4000, "API Server port")
	flag.StringVar(&conf.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&conf.db.dsn, "db-dsn", conf.vars.dbDSN, "data source name")

	// Read the connection pool settings from command-line flags into the config struct.
	// Notice the default values that we're using?
	flag.IntVar(&conf.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&conf.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&conf.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.Float64Var(&conf.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&conf.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&conf.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	// Read the redis configration setting into the config struct
	flag.StringVar(&conf.redis.host, "redis-host", conf.vars.redisHost, "redis host string")
	flag.StringVar(&conf.redis.port, "redis-port", conf.vars.redisPort, "redis port")
	flag.StringVar(&conf.redis.username, "redis-username", "", "redis username string")
	flag.StringVar(&conf.redis.password, "redis-password", conf.vars.redisPassword, "redis password string")
	// Create a new version boolean flag with the default value of false.
	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	// If the version flag value is true, then print out the version number and
	// immediately exit.
	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		fmt.Printf("Build time:\t%s\n", buildTime)
		os.Exit(0)
	}

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	rdb, err := openRedis(conf.redis.host, conf.redis.port, conf.redis.username, conf.redis.password)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer rdb.Close()

	logger.PrintInfo("redis connection pool established", nil)

	db, err := openDB(conf)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	// Defer a call to db.Close() so that the connection pool is closed before the
	// main() function exits.
	defer db.Close()

	logger.PrintInfo("database connection pool established", nil)

	// rabbitConn, err := connectAMQP(10, 1*time.Second)

	if err != nil {
		log.Fatal(err)
	}
	// defer rabbitConn.Close()
	logger.PrintInfo("rabbitmq connection established", nil)

	expvar.NewString("version").Set(version)

	// Publish the number of active goroutines.
	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))

	// Publish the database connection pool statistics.
	expvar.Publish("database", expvar.Func(func() interface{} {
		return db.Stats()
	}))

	// Publish the current Unix timestamp.
	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))

	cld, err := cloudinary.NewFromParams(conf.vars.cldName, conf.vars.cldAPIKey, conf.vars.cldSecret)
	if err != nil {
		log.Fatal(err)
	}
	app := &application{
		config: conf,
		logger: logger,
		models: data.NewModels(db, rdb),
		cld:    cld,
		// amqp:   rabbitConn,
	}
	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

func openDB(conf config) (*sql.DB, error) {
	db, err := sql.Open("postgres", conf.db.dsn)

	if err != nil {
		return nil, err
	}

	// Set the maximum number of open (in-use + idle) connections in the pool. Note that
	// passing a value less than or equal to 0 will mean there is no limit.
	db.SetMaxOpenConns(conf.db.maxOpenConns)

	// Set the maximum number of idle connections in the pool. Again, passing a value
	// less than or equal to 0 will mean there is no limit.
	db.SetMaxIdleConns(conf.db.maxIdleConns)

	// Use the time.ParseDuration() function to convert the idle timeout duration string
	// to a time.Duration type.
	duration, err := time.ParseDuration(conf.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	// Set the maximum idle timeout.
	db.SetConnMaxIdleTime(duration)

	// Create a context with a 5-second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use PingContext() to establish a new connection to the database, passing in the
	// context we created above as a parameter. If the connection couldn't be
	// established successfully within the 5 second deadline, then this will return an
	// error.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	// Return the sql.DB connection pool.
	return db, nil
}

func openRedis(host, port, username, password string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Password: "",
		Addr:     fmt.Sprint(host, ":", port),
		DB:       0,
		//Username: username,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := rdb.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}
	return rdb, nil
}

func loadEnvVars(conf *config) {
	conf.vars.dbDSN = os.Getenv("GREENLIGHT_DB_DSN")
	conf.vars.greenlightUserTokenCookie = os.Getenv("GREENLIGHT_TOKEN")
	conf.vars.greenlightUserIDCookie = os.Getenv("GREENLIGHT_USERID_TOKEN")
	conf.vars.emailPassword = os.Getenv("EMAIL_PASSWORD")
	conf.vars.clientUrl = os.Getenv("CLIENT_URL")
	if conf.vars.clientUrl == "" {
		conf.vars.clientUrl = "http://localhost:3000"
	}
	conf.vars.redisHost = os.Getenv("REDIS_HOST")
	if conf.vars.redisHost == "" {
		conf.vars.redisHost = "localhost"
	}
	conf.vars.redisPort = os.Getenv("REDIS_PORT")
	conf.vars.greenlightEmail = os.Getenv("GREENLIGHT_EMAIL")
	if conf.vars.greenlightEmail == "" {
		conf.vars.greenlightEmail = "no_replay@greenlight.com"
	}
	conf.vars.redisPassword = os.Getenv("REDIS_PASSWORD")
	conf.vars.cldName = os.Getenv("CLD_NAME")
	conf.vars.cldAPIKey = os.Getenv("CLD_API_KEY")
	conf.vars.cldSecret = os.Getenv("CLD_SECRET")
}

func connectAMQP(counts int64, backOff time.Duration) (*amqp.Connection, error) {
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
			return nil, fmt.Errorf("Can not connect to the RabbitMQ")
		}
		backOff = backOff + (time.Second * 2)

		fmt.Println("Backing off.....")
		time.Sleep(backOff)
		continue

	}
	return connection, nil
}
