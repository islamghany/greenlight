package main

import (
	"context"
	"net/http"

	"islamghany.greenlight/userspb"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *application) contextSetUser(r *http.Request, payload *userspb.AuthenticateResponse) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, payload)

	return r.WithContext(ctx)
}

func (app *application) contextGetUser(r *http.Request) *userspb.AuthenticateResponse {
	payload, ok := r.Context().Value(userContextKey).(*userspb.AuthenticateResponse)

	if !ok {
		panic("missing user value in request context")
	}

	return payload
}
