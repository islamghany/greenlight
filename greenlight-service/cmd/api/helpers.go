package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	//"islamghany.greenlight/internals/event"
	"islamghany.greenlight/internals/validator"

	"github.com/julienschmidt/httprouter"
)

// Define an envelope type.
type envelope map[string]interface{}

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)

	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

func (app *application) writeJsonString(w http.ResponseWriter, data string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, data)

}
func (app *application) writeJson(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	jsonData, err := json.MarshalIndent(data, "", "\t")

	if err != nil {
		return err
	}

	jsonData = append(jsonData, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonData)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dest interface{}) error {

	maxBytes := 1_048_576 // 1mg
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dest)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError): // dest is a nil pointer
			panic(err)

		// For anything else, return the error message as-is.
		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func (app *application) readString(q url.Values, key string, defaultValue string) string {

	s := q.Get(key)

	if s == "" {
		return defaultValue
	}
	return s
}

func (app *application) readInt(q url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := q.Get(key)
	if s == "" {
		return defaultValue
	}
	num, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be number")
		return defaultValue
	}
	return num

}
func (app *application) readCSV(q url.Values, key string, defaultValue []string) []string {
	s := q.Get(key)

	if s == "" {
		return defaultValue
	}
	return strings.Split(s, ",")

}

func (app *application) background(fn func()) {
	// Increment the WaitGroup counter.
	app.wg.Add(1)
	go func() {
		// Use defer to decrement the WaitGroup counter before the goroutine returns.
		defer app.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()
		fn()
	}()
}

func (app *application) contains(elems []string, v string) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

// func (app *application) addUserCookies(w http.ResponseWriter, token *data.Token) {
// 	app.addCookies(w, app.config.vars.greenlightUserTokenCookie, token.Plaintext, token.Expiry)
// 	app.addCookies(w, app.config.vars.greenlightUserIDCookie, fmt.Sprint(token.UserID), token.Expiry)
// }

// func (app *application) removeUsersCookies(w http.ResponseWriter) {
// 	app.removeCookies(w, app.config.vars.greenlightUserTokenCookie)
// 	app.removeCookies(w, app.config.vars.greenlightUserIDCookie)
// }

// func (app *application) pushToQueue(name, msg string) error {
// 	e, err := event.NewEventEmitter(app.amqp)

// 	if err != nil {
// 		return err
// 	}

// 	err = e.Push(msg, "mail.SEND")

// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
