package main

import (
	"errors"
	"net/http"
	"time"

	"islamghany.greenlight/internals/data"
	"islamghany.greenlight/internals/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		// If we get a ErrDuplicateEmail error, use the v.AddError() method to manually
		// add a message to the validator instance, and then call our
		// failedValidationResponse() helper.
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.models.Permissions.AddForUser(user.ID, "movies:read")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// After the user record has been created in the database, generate a new activation
	// token for the user.
	_, err = app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Use the background helper to execute an anonymous function that sends the welcome
	// email.
	// app.background(func() {
	// 	s := `
	// 	{
	// 		"from":"%s",
	// 		"to": "%s",
	// 		"data":{
	// 			"subject":"Activate your account",
	// 			"userID":%d,
	// 			"tokenExpirationTime":"3 days",
	// 			"activationToken":"%s/activate-account/%s"

	// 		},
	// 		"template_file":"user_welcome.tmpl"

	// 	}
	// 	`
	// 	jsonData := fmt.Sprintf(s,
	// 		app.config.vars.greenlightEmail,
	// 		user.Email,
	// 		user.ID,
	// 		app.config.vars.clientUrl,
	// 		token.Plaintext)

	// 	err = app.pushToQueue("name", jsonData)
	// 	if err != nil {
	// 		app.logger.PrintError(err, nil)
	// 	}
	// })

	//app.addUserCookies(w, token)
	err = app.writeJson(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Validate the plaintext token provided by the client.
	v := validator.New()
	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Retrieve the details of the user associated with the token using the
	// GetForToken() method (which we will create in a minute). If no matching record
	// is found, then we let the client know that the token they provided is not valid.
	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Update the user's activation status.
	user.Activated = true

	// Save the updated user record in our database, checking for any edit conflicts in
	// the same way that we did for our movie records.
	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// If everything went successfully, then we delete all activation tokens for the
	// user.
	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// This user is already verfied then he can write to any thing.
	// adding the write permission
	err = app.models.Permissions.AddForUser(user.ID, "movies:write")

	// Send the updated user details to the client in a JSON response.
	err = app.writeJson(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
func (app *application) getCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	err := app.writeJson(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	currentUserID := int64(-1)

	currentUser := app.contextGetUser(r)

	if !currentUser.IsAnonymous() {
		currentUserID = currentUser.ID
	}

	isCached := true
	jsonUser, err := app.models.Users.CacheRetrieveUserByID(id)

	if err != nil {
		isCached = false
	}

	if !isCached {
		user, err := app.models.Users.GetByID(id, currentUserID)

		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notFoundResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// cache in background
		app.background(func() {
			err = app.models.Users.CacheUserbyID(user.ID, envelope{"user": user})
			if err != nil {
				app.logger.PrintError(err, nil)
			}
		})

		err = app.writeJson(w, http.StatusOK, envelope{"user": user}, nil)
		if err != nil {
			app.serverErrorResponse(w, r, err)
		}
	} else {
		app.writeJsonString(w, *jsonUser)
	}

}
func (app *application) signedUserOutHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	err := app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	app.removeUsersCookies(w)

	err = app.writeJson(w, http.StatusOK, envelope{"message": "signed out successfully."}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
