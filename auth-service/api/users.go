package api

import (
	db "auth-service/db/sqlc"
	"auth-service/utils"
	"context"
	"net/http"
)

var AnonymousUser = &db.User{}

func IsAnonymous(u *db.User) bool {
	return u == AnonymousUser
}

func (server *Server) registerUserHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Name     string `json:"name" validate:"required,min=6,max=72"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=6,max=72"`
	}
	err := server.readJSON(w, r, &input)
	if err != nil {
		server.badRequestResponse(w, r, err)
		return
	}

	err = server.validator.V.Struct(input)

	if err != nil {

		server.validationErrorResponse(w, r, err, server.validator)
		return
	}
	user := db.CreateUserParams{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	hash, err := utils.HashPassword(input.Password)
	if err != nil {
		server.serverErrorResponse(w, r, err)
		return
	}

	user.HashedPassword = hash

	u, err := server.store.CreateUser(context.Background(), user)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			server.errorResponse(w, r, http.StatusUnprocessableEntity, "this email is alreay exists")
		default:
			server.serverErrorResponse(w, r, err)
		}
		return
	}
	err = server.writeJson(w, http.StatusCreated, envelope{"user": u}, nil)
	if err != nil {
		server.serverErrorResponse(w, r, err)
	}
}
