package api

import (
	"logger-service/logspb"

	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	db *mongo.Client
	logspb.UnimplementedLogServiceServer
}

func NewServer(db *mongo.Client) *Server {

	return &Server{
		db: db,
	}

}
