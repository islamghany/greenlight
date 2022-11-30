package api

import (
	"auth-service/db/cache"
	db "auth-service/db/sqlc"
	"auth-service/token"
	"auth-service/utils"
	"fmt"
	"net/http"
	"time"
	//"gopkg.in/go-playground/validator.v9"
)

type envelope map[string]interface{}

type Server struct {
	store     *db.Queries
	cache     *cache.Cache
	config    *utils.Config
	validator *utils.UserValidtor
	maker     token.Maker
}

func NewServer(s *db.Queries, c *cache.Cache, conf *utils.Config, v *utils.UserValidtor, maker token.Maker) *Server {
	return &Server{
		store:     s,
		cache:     c,
		config:    conf,
		validator: v,
		maker:     maker,
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

	return srv.ListenAndServe()
}
