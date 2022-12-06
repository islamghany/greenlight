package api

import (
	"mailer-service/mailer"
	"mailer-service/mailpb"
)

type Server struct {
	mailer mailer.Mailer
	mailpb.UnimplementedMailSeviceServer
}

func NewServer(m mailer.Mail) (*Server, error) {

	return &Server{
		mailer: *mailer.NewMailer(m),
	}, nil

}
