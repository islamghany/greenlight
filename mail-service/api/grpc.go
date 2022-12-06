package api

import (
	"context"
	"errors"
	"fmt"
	"mailer-service/mailpb"
	"net"

	"google.golang.org/grpc"
)

func (s *Server) OpenGRPC(port int) error {

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))

	if err != nil {
		return err
	}

	gServer := grpc.NewServer()
	mailpb.RegisterMailSeviceServer(gServer, s)

	if err := gServer.Serve(lis); err != nil {
		return err
	}
	return nil
}

func (server *Server) InsertLog(ctx context.Context, req *mailpb.MailRequest) (*mailpb.MailResponse, error) {

	m := req.GetMailEntry()

	if m.TemplateFile == "" {
		return nil, errors.New("template File must be a valid string path")
	}

	err := server.mailer.Send(m)
	if err != nil {
		return nil, err
	}
	res := &mailpb.MailResponse{Message: "done"}

	return res, nil

}
