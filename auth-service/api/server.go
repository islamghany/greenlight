package api

import (
	"auth-service/db/cache"
	db "auth-service/db/sqlc"
	"auth-service/event"
	"auth-service/token"
	"auth-service/userspb"
	"auth-service/utils"
	"fmt"
	"log"
	"net/http"
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

func (server *Server) Start(port int) {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      server.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
