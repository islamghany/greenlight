package api

import (
	"fmt"
	"net/http"
	"time"
)

type envelope map[string]interface{}
type Server struct {
}

func NewServer() (*Server, error) {

	return &Server{}, nil

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
