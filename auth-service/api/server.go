package api

import (
	"auth-service/db/cache"
	db "auth-service/db/sqlc"
	"auth-service/event"
	"auth-service/logspb"
	"auth-service/token"
	"auth-service/userspb"
	"auth-service/utils"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type envelope map[string]interface{}

type Server struct {
	store     *db.Queries
	cache     *cache.Cache
	config    *utils.Config
	validator *utils.UserValidtor
	maker     token.Maker
	userspb.UnimplementedUserServiceServer
	amqp    *amqp.Connection
	emitter *event.Emitter
	wg      sync.WaitGroup
}

func NewServer(
	s *db.Queries,
	c *cache.Cache,
	conf *utils.Config,
	v *utils.UserValidtor,
	maker token.Maker,
	amqp *amqp.Connection,
	emitter *event.Emitter,
) *Server {
	return &Server{
		store:     s,
		cache:     c,
		config:    conf,
		validator: v,
		maker:     maker,
		amqp:      amqp,
		emitter:   emitter,
	}
}

func (server *Server) Start(port int) error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      server.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownError := make(chan error)

	// intercepting the shutdown
	go func() {
		quit := make(chan os.Signal, 1)

		// Use signal.Notify() to listen for incoming SIGINT and SIGTERM signals and
		// relay them to the quit channel. Any other signals will not be caught by
		// signal.Notify() and will retain their default behavior.
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// Read the signal from the quit channel. This code will block until a signal is
		// received.
		s := <-quit

		// Log a message to say that the signal has been caught. Notice that we also
		// call the String() method on the signal to get the signal name and include it
		// in the log entry properties.
		log.Printf("shutting down server, signal: %s", s.String())

		err := server.emitter.SendToLogService(&logspb.Log{
			ErrorMessage: "shutting down server",
			StackTrace:   fmt.Sprintf("signal %s", s.String()),
		})

		if err != nil {
			log.Println(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Call Shutdown() on the server like before, but now we only send on the
		// shutdownError channel if it returns an error.
		// Call Shutdown() on our server, passing in the context we just made.
		// Shutdown() will return nil if the graceful shutdown was successful, or an
		// error (which may happen because of a problem closing the listeners, or
		// because the shutdown didn't complete before the 5-second context deadline is
		// hit). We relay this return value to the shutdownError channel.
		err = srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		// Log a message to say that we're waiting for any background goroutines to
		// complete their tasks.
		log.Printf("completing background tasks, addr: %s", srv.Addr)

		// Call Wait() to block until our WaitGroup counter is zero --- essentially
		// blocking until the background goroutines have finished. Then we return nil on
		// the shutdownError channel, to indicate that the shutdown completed without
		// any issues.
		server.wg.Wait()
		shutdownError <- nil
	}()
	log.Printf(`Starting Server
		"addr": %s,
	`, srv.Addr)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Otherwise, we wait to receive the return value from Shutdown() on the
	// shutdownError channel. If return value is an error, we know that there was a
	// problem with the graceful shutdown and we return the error.
	err = <-shutdownError
	if err != nil {
		return err
	}

	// At this point we know that the graceful shutdown completed successfully and we
	// log a "stopped server" message.
	log.Printf(`Stopped Server
		"addr": %s,
	`, srv.Addr)

	return nil
}
