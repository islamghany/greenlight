package api

import (
	"auth-service/db/cache"
	db "auth-service/db/sqlc"
	"auth-service/utils"
	"fmt"
	"net/http"
	"time"
)

type envelope map[string]interface{}

type Server struct {
	store  *db.Queries
	cache  *cache.Cache
	config *utils.Config
}

func NewServer(s *db.Queries, c *cache.Cache, conf *utils.Config) *Server {
	return &Server{
		store:  s,
		cache:  c,
		config: conf,
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
