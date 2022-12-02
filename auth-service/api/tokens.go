package api

import (
	db "auth-service/db/sqlc"
	"auth-service/utils"
	"context"
	"database/sql"
	"net/http"
	"time"

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

	server.addUserCookies(w, accessToken, refreshToken, accessPayload.ExpiredAt, refreshPayload.ExpiredAt)

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

func (server *Server) renewAccessTokenHandler(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("refresh_token")

	if err != nil || cookie.Value == "" {
		server.authenticationRequiredResponse(w, r)
		return
	}

	payload, err := server.maker.VerifyToken(cookie.Value)

	if err != nil {
		server.invalidAuthenticationTokenResponse(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session, err := server.store.GetSession(ctx, payload.ID)

	if err != nil {
		server.serverErrorResponse(w, r, err)
		return
	}

	isAborted := false

	if session.RefreshToken != cookie.Value {
		server.removeUserCookies(w)
		isAborted = true
		server.invalidAuthenticationTokenResponse(w, r)
	}

	if !isAborted && time.Now().After(session.ExpiresAt) {
		server.removeUserCookies(w)
		isAborted = true
		server.invalidAuthenticationTokenResponse(w, r)
	}

	if isAborted {
		server.background(func() {
			server.store.DeleteAllSessionForUser(context.Background(), session.UserID)
		})
		return
	}

	accessToken, accessPayload, err := server.maker.CreateToken(session.UserID, server.config.ACCESS_TOKEN_DURATION)

	if err != nil {
		server.serverErrorResponse(w, r, err)
		return
	}

	server.setAccessTokenCookie(w, accessToken, accessPayload.ExpiredAt)

	err = server.writeJson(w, http.StatusCreated, envelope{
		"access_token_expires_at": accessPayload.ExpiredAt,
	}, nil)

	if err != nil {
		server.serverErrorResponse(w, r, err)
		return
	}
}
