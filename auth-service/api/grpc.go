package api

import (
	"auth-service/userspb"
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
)

func (s *Server) OpenGRPC(port int) error {

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))

	if err != nil {
		return err
	}

	gServer := grpc.NewServer()
	userspb.RegisterUserServiceServer(gServer, s)

	if err := gServer.Serve(lis); err != nil {
		return err
	}
	return nil
}

func (server *Server) Authenticate(ctx context.Context, req *userspb.AuthenticateRequst) (*userspb.AuthenticateResponse, error) {

	input := req.GetAccessToken()

	payload, err := server.maker.VerifyToken(input)

	if err != nil {
		return nil, err
	}

	user, err := server.store.GetUserByID(ctx, payload.UserID)

	if err != nil {
		return nil, err
	}

	permissions, err := server.store.GetAllPermissionsForUser(ctx, payload.UserID)

	if err != nil {
		return nil, err
	}

	return &userspb.AuthenticateResponse{
		User: &userspb.User{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Name:      user.Name,
			Activated: user.Activated,
		},
		Permissions: permissions,
	}, nil

}

func (server *Server) GetUser(ctx context.Context, req *userspb.AuthenticateRequst) (*userspb.User, error) {
	input := req.GetAccessToken()
	payload, err := server.maker.VerifyToken(input)

	if err != nil {
		return nil, err
	}
	user, err := server.store.GetUserByID(ctx, payload.UserID)

	if err != nil {
		return nil, err
	}

	return &userspb.User{
		Id:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Name:      user.Name,
		Activated: user.Activated,
	}, nil
}
