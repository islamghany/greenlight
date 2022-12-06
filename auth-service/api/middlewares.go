package api

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func (server *Server) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event of a panic
		// as Go unwinds the stack).
		defer func() {
			// Use the builtin recover function to check if there has been a panic or
			// not.
			if err := recover(); err != nil {
				// If there was a panic, set a "Connection: close" header on the
				// response. This acts as a trigger to make Go's HTTP server
				// automatically close the current connection after a response has been
				// sent.
				w.Header().Set("Connection", "close")
				// The value returned by recover() has the type interface{}, so we use
				// fmt.Errorf() to normalize it into an error and call our
				// serverErrorResponse() helper. In turn, this will log the error using
				// our custom Logger type at the ERROR level and send the client a 500
				// Internal Server Error response.
				server.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (server *Server) authenticateMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("access_token")

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

		user, err := server.store.GetUserByID(ctx, payload.UserID)

		if err != nil {
			server.serverErrorResponse(w, r, err)
			return
		}

		r = server.contextSetUser(r, &user)

		next.ServeHTTP(w, r)
	}
}
