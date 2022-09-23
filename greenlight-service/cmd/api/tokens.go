package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"islamghany.greenlight/internals/data"

	"islamghany.greenlight/internals/validator"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Validate the email and password provided by the client.
	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}
	// Otherwise, if the password is correct, we generate a new token with a 24-hour
	// expiry time and the scope 'authentication'.
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.addUserCookies(w, token)

	// Encode the token to JSON and send it in the response along with a 201 Created
	// status code.
	err = app.writeJson(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createResetPasswordTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()

	if data.ValidateEmail(v, input.Email); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	user, err := app.models.Users.GetByEmail(input.Email)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	token, err := app.models.Tokens.New(user.ID, 30*time.Minute, data.ScopePasswordRecovery)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.background(func() {
		s := `
		{
			"from":"%s",
			"to": "%s",
			"data":{
				"subject":"Reset your password",
				"userID":%d,
				"tokenExpirationTime":"30 Minutes",
				"clientUrl":"%s/reset-password/%s"
				
			},
			"template_file":"reset_password.temp"
			
		}
		`
		jsonData := fmt.Sprintf(s,
			app.config.vars.greenlightEmail,
			user.Email,
			user.ID,
			app.config.vars.clientUrl,
			token.Plaintext)

		app.logger.PrintInfo(jsonData, nil)
		err = app.pushToQueue("name", jsonData)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})

	err = app.writeJson(w, http.StatusCreated, envelope{"message": "an mail was sent to your email address please follow instruction so that you can reset your password."}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
