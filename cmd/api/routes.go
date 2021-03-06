package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/movies", app.listMoviesHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requirePermission("movies:write", app.createMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requirePermission("movies:write", app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requirePermission("movies:write", app.deleteMovieHandler))
	router.HandlerFunc(http.MethodOptions, "/v1/movies/most-likes", app.GetMostLikedMovivesHandler)
	router.HandlerFunc(http.MethodOptions, "/v1/movies/most-views", app.GetMostViewsMovivesHandler)

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodGet, "/v1/users", app.requireAuthenticatedUser(app.getCurrentUserHandler))
	router.HandlerFunc(http.MethodPost, "/v1/users/signout", app.requireAuthenticatedUser(app.signedUserOutHandler))
	router.HandlerFunc(http.MethodGet, "/v1/users/:id", app.getUserHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/reset-password-token", app.createResetPasswordTokenHandler)

	router.HandlerFunc(http.MethodGet, "/v1/likes/:id", app.getMoiveLikeHandler)
	router.HandlerFunc(http.MethodPost, "/v1/likes", app.requirePermission("movies:write", app.addLikeHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/likes/:id", app.requirePermission("movies:write", app.deleteLikeHandler))

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	// Wrap the router with the panic recovery middleware.
	return app.metrics(app.recoverPanic(app.enableCORs(app.cacheReinvalidate(app.rateLimiter(app.authenticateViaCookie(router))))))
}
