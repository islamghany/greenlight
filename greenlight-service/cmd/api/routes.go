package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"expvar"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
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

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/send-email", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("heeelelfmg")
		var msg MailPayload

		err := app.readJSON(w, r, &msg)
		jsonData, _ := json.MarshalIndent(msg, "", "\t")
		if err != nil {
			app.badRequestResponse(w, r, err)
			return
		}

		// call the mail service
		mailServiceURL := "http://mail-service/send"

		// post to mail service
		request, err := http.NewRequest("POST", mailServiceURL, bytes.NewBuffer(jsonData))
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		request.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		response, err := client.Do(request)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		defer response.Body.Close()

		// make sure we get back the right status code
		if response.StatusCode != http.StatusCreated {
			app.serverErrorResponse(w, r, errors.New("error calling mail service"))
			return
		}

		// send back json

		err = app.writeJson(w, http.StatusCreated, envelope{"user": msg.To}, nil)
		if err != nil {
			app.serverErrorResponse(w, r, err)
		}

	})
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
