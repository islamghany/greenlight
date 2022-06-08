package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"islamghany.greenlight/internals/data"
	"islamghany.greenlight/internals/jsonlog"
	"islamghany.greenlight/internals/mailer"
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
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	redis struct {
		host     string
		port     string
		username string
		password string
	}
}

// app struct to hold the http handlers, helpers and middleware
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
	rdb    *redis.Client
}

func main() {
	var conf config

	flag.IntVar(&conf.port, "port", 4000, "API Server port")
	flag.StringVar(&conf.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&conf.db.dsn, "db-dsn", "", "data source name")

	// Read the connection pool settings from command-line flags into the config struct.
	// Notice the default values that we're using?
	flag.IntVar(&conf.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&conf.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&conf.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.Float64Var(&conf.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&conf.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&conf.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	// Read the SMTP server configuration settings into the config struct, using the
	// Mailtrap settings as the default values. IMPORTANT: If you're following along,
	// make sure to replace the default values for smtp-username and smtp-password
	// with your own Mailtrap credentials.
	flag.StringVar(&conf.smtp.host, "smtp-host", "smtp.gmail.com", "SMTP host")
	flag.IntVar(&conf.smtp.port, "smtp-port", 587, "SMTP port")
	flag.StringVar(&conf.smtp.username, "smtp-username", "dump.dumper77@gmail.com", "SMTP email")
	flag.StringVar(&conf.smtp.password, "smtp-password", os.Getenv("EMAIL_PASSWORD"), "SMTP password")
	flag.StringVar(&conf.smtp.sender, "smtp-sender", "dump.dumper77@gmail.com", "SMTP sender")

	// Read the redis configration setting into the config struct
	flag.StringVar(&conf.redis.host, "redis-host", os.Getenv("REDIS_HOST"), "redis host string")
	flag.StringVar(&conf.redis.port, "redis-port", os.Getenv("REDIS_PORT"), "redis port")
	flag.StringVar(&conf.redis.username, "redis-username", "", "redis username string")
	flag.StringVar(&conf.redis.password, "redis-password", os.Getenv("REDIS_PASSWORD"), "redis password string")
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

	app := &application{
		config: conf,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(conf.smtp.host, conf.smtp.port, conf.smtp.username, conf.smtp.password, conf.smtp.sender)}
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
		Password: password,
		Addr:     fmt.Sprint(host, ":", port),
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
