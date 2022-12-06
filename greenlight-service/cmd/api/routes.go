package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"islamghany.greenlight/userspb"
)

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/movies", app.listMoviesHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.authenticate(app.requirePermission(userspb.PERMISSION_CODE_movies_write.String(), app.createMovieHandler)))
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.authenticate(app.requirePermission(userspb.PERMISSION_CODE_movies_write.String(), app.showMovieHandler)))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.authenticate(app.requirePermission(userspb.PERMISSION_CODE_movies_write.String(), app.updateMovieHandler)))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.authenticate(app.requirePermission(userspb.PERMISSION_CODE_movies_write.String(), app.deleteMovieHandler)))
	router.HandlerFunc(http.MethodGet, "/v1/most-movies/likes", app.GetMostLikedMovivesHandler)
	router.HandlerFunc(http.MethodGet, "/v1/most-movies/views", app.GetMostViewsMovivesHandler)

	router.HandlerFunc(http.MethodGet, "/v1/likes/:id", app.getMoiveLikeHandler)
	router.HandlerFunc(http.MethodPost, "/v1/likes", app.authenticate(app.requirePermission(userspb.PERMISSION_CODE_movies_write.String(), app.addLikeHandler)))
	router.HandlerFunc(http.MethodDelete, "/v1/likes/:id", app.authenticate(app.requirePermission(userspb.PERMISSION_CODE_movies_write.String(), app.deleteLikeHandler)))

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	// Wrap the router with the panic recovery middleware.
	return app.metrics(app.recoverPanic(app.enableCORs(app.cacheReinvalidate(app.rateLimiter(router)))))
}
