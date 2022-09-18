package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (server *Server) routes() http.Handler {

	router := httprouter.New()

	router.HandlerFunc(http.MethodPost, "/send", server.SendMail)
	return server.enableCORS(router)
}
