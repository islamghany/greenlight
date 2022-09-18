package api

import (
	"fmt"
	"mailer-service/mailer"
	"net/http"
	"time"
)

type envelope map[string]interface{}
type Server struct {
	mailer mailer.Mailer
}

func NewServer(m mailer.Mail) (*Server, error) {

	return &Server{
		mailer: *mailer.NewMailer(m),
	}, nil

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
