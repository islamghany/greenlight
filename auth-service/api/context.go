package api

import (
	db "auth-service/db/sqlc"
	"context"
	"net/http"
)

type contextKey string

const userContextKey = contextKey("user")

func (server *Server) contextSetUser(r *http.Request, user *db.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (server *Server) contextGetUser(r *http.Request) *db.User {
	user, ok := r.Context().Value(userContextKey).(*db.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
