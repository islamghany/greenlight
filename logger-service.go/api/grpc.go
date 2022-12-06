package api

import (
	"context"
	"fmt"
	"logger-service/logspb"
	"net"
	"time"

	"google.golang.org/grpc"
)

type LogEntry struct {
	ID            string    `bson:"_id,omitempty" json:"id,omitempty"`
	ErrorMessagee string    `bson:"error_message" json:"error_message"`
	ServiceName   string    `bson:"service_name" json:"service_name"`
	StackTrace    string    `bson:"stack_trace" json:"stack_trace"`
	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
}

func (s *Server) OpenGRPC(port int) error {

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))

	if err != nil {
		return err
	}

	gServer := grpc.NewServer()
	logspb.RegisterLogServiceServer(gServer, s)

	if err := gServer.Serve(lis); err != nil {
		return err
	}
	return nil
}

func (server *Server) InsertLog(ctx context.Context, req *logspb.LogRequest) (*logspb.LogResponse, error) {

	l := req.GetLog()
	fmt.Println("Log come,", l)
	collection := server.db.Database("logs").Collection("logs")

	_, err := collection.InsertOne(context.TODO(), LogEntry{
		ServiceName:   l.ServiceName,
		StackTrace:    l.StackTrace,
		ErrorMessagee: l.ErrorMessage,
		CreatedAt:     time.Now(),
	})
	if err != nil {
		fmt.Println("Error inserting into logs:", err)
		return nil, err
	}

	res := &logspb.LogResponse{Message: "done"}

	return res, nil

}
