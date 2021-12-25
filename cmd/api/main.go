package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// version of the application
const version = "1.0.0"

// config
type config struct {
	port int    // port number e.g. 8080
	env  string // development ot prodction
}

// app struct to hold the http handlers, helpers and middleware
type application struct {
	config config
	logger *log.Logger
}

func main() {
	var conf config

	flag.IntVar(&conf.port, "port", 4000, "API Server port")
	flag.StringVar(&conf.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	app := &application{
		config: conf,
		logger: logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello from the other side\n")
	})
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", conf.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	logger.Printf("Server is running on port %d in %s mode", app.config.port, conf.env)

	err := srv.ListenAndServe()
	if err != nil {
		logger.Fatal(err)
	}
}
