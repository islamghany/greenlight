package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (server *Server) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(server.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(server.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/accounts/users", server.authenticateMiddleware(server.getCurrentUserHandler))
	router.HandlerFunc(http.MethodPost, "/v1/accounts/users", server.registerUserHandler)
	router.HandlerFunc(http.MethodGet, "/v1/accounts/users/:id", server.getUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/accounts/users/logout", server.authenticateMiddleware(server.logoutHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/accounts/users/activate", server.activateUserHandler)

	router.HandlerFunc(http.MethodPost, "/v1/accounts/tokens/authentication", server.createAuthenticationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/v1/accouts/tokens/renew-access-token", server.renewAccessTokenHandler)
	return server.recoverPanic(router)
}
