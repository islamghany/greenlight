package api

import (
	"context"
	"net/http"
	"time"
)

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
