package api

import (
	"context"
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
	mailpb.RegisterMailSeviceServer(gServer, &Server{})

	if err := gServer.Serve(lis); err != nil {
		return err
	}
	return nil
}

func (server *Server) SendMail(ctx context.Context, req *mailpb.MailRequest) (*mailpb.MailResponse, error) {
	// data := map[string]interface{}{
	// 	"subject": "Say my name",
	// 	"message": "heisenbsserg",
	// }
	// msg := mailer.Message{
	// 	From:         "me@example.com",
	// 	To:           "islamghany3@gmail.com",
	// 	Data:         data,
	// 	TemplateFile: "user_welcome.tmpl",
	// }
	// server.background(func() {
	// 	err := server.mailer.Send(msg)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}

	// })
	msg := Message{

		From:    "me@example.com",
		To:      "islamghany3@gmail.com",
		Data:    "ndfbfbdhfbhdf",
		Subject: "islam",
	}

	err := server.Mailer.SendSMTPMessage(msg)
	if err != nil {
		fmt.Println(err)
	}
	// input := req.GetMailEntry()

	// fmt.Println(input)

	// data := map[string]interface{}{
	// 	"subject": input.Subject,
	// 	"message": input.Message,
	// }

	// msg := mailer.Message{
	// 	From:         input.From,
	// 	To:           input.To,
	// 	Data:         data,
	// 	TemplateFile: "user_welcome.tmpl",
	// 	Attachments:  input.Attachment,
	// }
	// fmt.Println("2 :")
	// s.background(func() {
	// 	err := s.mailer.Send(msg)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}

	// })
	// // err := s.mailer.Send(msg)
	// // fmt.Println("3 :")
	// // if err != nil {
	// // 	res := &mailpb.MailResponse{Message: "failed!"}
	// // 	return res, err
	// // }
	// // fmt.Println("3 :")

	res := &mailpb.MailResponse{Message: "d"}

	return res, nil

}
