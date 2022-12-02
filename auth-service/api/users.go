package api

import (
	db "auth-service/db/sqlc"
	"auth-service/userspb"
	"auth-service/utils"
	"context"
	"database/sql"
	"net/http"
	"time"
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
		Username string `json:"username" validate:"required,min=6,max=72"`
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
		Username:  input.Username,
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
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			server.errorResponse(w, r, http.StatusUnprocessableEntity, "this username is alreay exists")
		default:
			server.serverErrorResponse(w, r, err)
		}
		return
	}
	err = server.store.AddPermissionForUser(context.Background(), db.AddPermissionForUserParams{
		ID:   u.ID,
		Code: userspb.PERMISSION_CODE_movies_read.String(),
	})

	if err != nil {
		server.serverErrorResponse(w, r, err)
		return
	}
	err = server.writeJson(w, http.StatusCreated, envelope{"user": u}, nil)
	if err != nil {
		server.serverErrorResponse(w, r, err)
	}
}

func (server *Server) getUserHandler(w http.ResponseWriter, r *http.Request) {

	id, err := server.readIDParams(r)
	if err != nil {
		server.notFoundResponse(w, r)
		return
	}

	user, err := server.store.GetUserByID(context.Background(), id)

	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			server.invalidCredentialsResponse(w, r)
		default:
			server.serverErrorResponse(w, r, err)
		}
		return
	}
	user.HashedPassword = nil
	err = server.writeJson(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		server.serverErrorResponse(w, r, err)
	}
}

func (server *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {

	// user must be authenticated to get into this handler.

	user := server.contextGetUser(r)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := server.store.DeleteAllSessionForUser(ctx, user.ID)

	if err != nil {
		server.serverErrorResponse(w, r, err)
		return
	}

	server.removeUserCookies(w)

	err = server.writeJson(w, http.StatusOK, envelope{"message": "signed out successfully."}, nil)
	if err != nil {
		server.serverErrorResponse(w, r, err)
	}
}
