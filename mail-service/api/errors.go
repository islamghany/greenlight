package api

import (
	"fmt"
	"net/http"
)

func (server *Server) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{"error": message}

	err := server.writeJson(w, status, env, nil)

	if err != nil {
		w.WriteHeader(500)
	}
}

func (server *Server) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	fmt.Println(err)
	message := "the server encountered a problem and could not process your request"
	server.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (server *Server) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	server.errorResponse(w, r, http.StatusBadRequest, err.Error())
}
