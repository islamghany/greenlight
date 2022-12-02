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
	router.HandlerFunc(http.MethodPost, "/v1/accounts/users", server.registerUserHandler)
	router.HandlerFunc(http.MethodGet, "/v1/accounts/users/:id", server.getUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/accounts/users/logout", server.authenticateMiddleware(server.logoutHandler))

	router.HandlerFunc(http.MethodPost, "/v1/accounts/tokens/authentication", server.createAuthenticationTokenHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/accoutns/tokens/logout", server.logoutHandler)

	return router
}
