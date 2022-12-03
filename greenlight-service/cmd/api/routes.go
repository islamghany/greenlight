package main

import (
	"encoding/json"
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"islamghany.greenlight/internals/event"
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

	//router.HandlerFunc(http.MethodGet, "/v1/movies/check/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies/test/send-email", func(w http.ResponseWriter, r *http.Request) {

		// defer r.Body.Close()

		// jsonData, err := io.ReadAll(r.Body)

		// if err != nil {
		// 	app.badRequestResponse(w, r, err)
		// 	return
		// }
		// fmt.Println(jsonData)
		payload := event.Payload{
			Name: "mail",
			Data: `{"name":islam, age:23}`,
		}
		msg, err := json.Marshal(payload)

		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		err = app.emitter.Push(string(msg), "mail")

		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		// err = app.pushToQueue("name", string(jsonData))

		err = app.writeJson(w, http.StatusCreated, envelope{"messgae": "message successfully sent to"}, nil)
		if err != nil {
			app.serverErrorResponse(w, r, err)
		}

	})
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
