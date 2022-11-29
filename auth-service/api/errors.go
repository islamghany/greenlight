package api

import (
	"fmt"
	"log"
	"net/http"
)

func (server *Server) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{"error": message}
	err := server.writeJson(w, status, env, nil)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
	}
}
func (server *Server) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Println("error: ", err)

	message := "the server encountered a problem and could not process your request"
	server.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (server *Server) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	server.errorResponse(w, r, http.StatusNotFound, message)
}

func (server *Server) forbiddenResponse(w http.ResponseWriter, r *http.Request) {
	message := "authentication is impossible for this user and browsers will not propose a new attempt."
	server.errorResponse(w, r, http.StatusForbidden, message)
}

func (server *Server) unauthorizedResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid credentials"
	server.errorResponse(w, r, http.StatusUnauthorized, message)
}
func (serve *Server) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	serve.errorResponse(w, r, http.StatusBadRequest, err.Error())
}
