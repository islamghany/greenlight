package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (server *Server) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/", func(w http.ResponseWriter, r *http.Request) {

		w.Write([]byte("hellow worrld"))
	})
	router.HandlerFunc(http.MethodPost, "/api/v1/users", server.registerUserHandler)

	return router
}
