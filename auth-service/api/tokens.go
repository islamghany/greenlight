package api

import (
	db "auth-service/db/sqlc"
	"auth-service/utils"
	"context"
	"database/sql"
	"net/http"

	"github.com/tomasen/realip"
)

func (server *Server) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
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

	user, err := server.store.GetUserByEmail(context.Background(), input.Email)

	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			server.invalidCredentialsResponse(w, r)
		default:
			server.serverErrorResponse(w, r, err)
		}
		return
	}

	err = utils.CheckPassword(input.Password, user.HashedPassword)

	if err != nil {
		server.invalidCredentialsResponse(w, r)
		return
	}

	accessToken, accessPayload, err := server.maker.CreateToken(user.ID, server.config.ACCESS_TOKEN_DURATION)

	if err != nil {
		server.serverErrorResponse(w, r, err)
		return
	}

	refreshToken, refreshPayload, err := server.maker.CreateToken(user.ID, server.config.REFRESH_TOKEN_DURATION)

	if err != nil {
		server.serverErrorResponse(w, r, err)
		return
	}

	session, err := server.store.InsertSession(context.Background(), db.InsertSessionParams{
		ID:           refreshPayload.ID,
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    r.UserAgent(),
		UserIp:       realip.FromRequest(r),
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		server.serverErrorResponse(w, r, err)
		return
	}

	server.setCooke(w, "access_token", accessToken, "/", accessPayload.ExpiredAt)
	server.setCooke(
		w,
		"refresh_token",
		refreshToken,
		"/v1/accouts/tokens/renew-access-token",
		refreshPayload.ExpiredAt)

	err = server.writeJson(w, http.StatusCreated, envelope{
		"user":                     user,
		"session_id":               session.ID,
		"access_token_expires_at":  accessPayload.ExpiredAt,
		"refresh_token_expires_at": refreshPayload.ExpiredAt,
	}, nil)

	if err != nil {
		server.serverErrorResponse(w, r, err)
		return
	}
}

func (server *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {

}
