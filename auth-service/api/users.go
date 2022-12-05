package api

import (
	db "auth-service/db/sqlc"
	"auth-service/mailpb"
	"auth-service/userspb"
	"auth-service/utils"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

var AnonymousUser = &db.User{}

func IsAnonymous(u *db.User) bool {
	return u == AnonymousUser
}

// TODO : add the register operations in a transaction.
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

	// after created user with the permission, now user must activate his account via his mail
	// so i will geneate an token and send it his email and then he can varify his account through it.

	activationToken, activationPayload, err := server.maker.CreateToken(u.ID, server.config.ACTIVATE_TOKEN_DURATION)
	if err != nil {
		server.serverErrorResponse(w, r, err)
		return
	}

	// I will send the a message to message broker to send it to the mail service.
	m := mailpb.Mail{
		From:         server.config.AUTH_EMAIL,
		To:           user.Email,
		Subject:      "Activate your account",
		TemplateFile: "user_welcome.tmpl",
		Data: map[string]string{
			"subject":             "Activate your account",
			"userID":              fmt.Sprintf("%d", u.ID),
			"tokenExpirationTime": fmt.Sprintf("%f minutes", time.Until(activationPayload.ExpiredAt).Minutes()),
			"activationToken":     fmt.Sprintf("user this token through /v1/accounts/user/activate, activation Token is: %s", activationToken),
		},
	}

	err = server.emitter.SendToMailService(&m)

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

	user, err := server.getReadyUser(id)

	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			server.notFoundResponse(w, r)
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

// TODO : add the activations operations in a transaction.

func (server *Server) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token" validate:"required"`
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

	activationPayload, err := server.maker.VerifyToken(input.TokenPlaintext)
	if err != nil {
		server.invalidAuthenticationTokenResponse(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	user, err := server.store.GetUserByID(ctx, activationPayload.UserID)

	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			server.notFoundResponse(w, r)
		default:
			server.serverErrorResponse(w, r, err)
		}
		return
	}

	cancel()

	updatedUser, err := server.store.UpdateUser(context.Background(), db.UpdateUserParams{
		ID:        user.ID,
		Version:   user.Version,
		Activated: sql.NullBool{Bool: true, Valid: true},
	})

	if err != nil {
		server.serverErrorResponse(w, r, err)
		return
	}

	// This user is already verfied then he can write to any thing.
	// adding the write permission

	err = server.store.AddPermissionForUser(context.Background(), db.AddPermissionForUserParams{
		ID:   user.ID,
		Code: userspb.PERMISSION_CODE_movies_write.String(),
	})

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_permissions_pkey"`:
			server.invalidAuthenticationTokenResponse(w, r)
		default:
			server.serverErrorResponse(w, r, err)
		}
		return
	}
	if err != nil {

		server.serverErrorResponse(w, r, err)
		return
	}

	err = server.writeJson(w, http.StatusOK, envelope{"user": updatedUser}, nil)
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
