package main

import (
	"errors"
	"net/http"

	"islamghany.greenlight/internals/data"
	"islamghany.greenlight/internals/validator"
)

func (app *application) addLikeHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		MovieID int64 `json:"movie_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// check if the current user id is equal to the user id that was sent.

	user := app.contextGetUser(r)

	v := validator.New()
	data.ValidateLikeInput(v, input.MovieID)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Likes.Insert(user.ID, input.MovieID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateLike):
			app.badRequestResponse(w, r, err)
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJson(w, http.StatusCreated, envelope{"message": "succfully added the like"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteLikeHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	err = app.models.Likes.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJson(w, http.StatusOK, envelope{"message": "like successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
