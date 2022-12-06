package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

func (server *Server) readIDParams(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)

	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

func (server *Server) writeJson(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
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

func (server *Server) readJSON(w http.ResponseWriter, r *http.Request, dest interface{}) error {

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

func (app *Server) background(fn func()) {
	// Increment the WaitGroup counter.
	app.wg.Add(1)
	go func() {

		// Use defer to decrement the WaitGroup counter before the goroutine returns.
		defer app.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()
		fn()
	}()
}

func (server *Server) setCookie(w http.ResponseWriter, name, value, path string, ttl time.Time) {

	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		HttpOnly: true,
		Expires:  ttl,
		Secure:   true,
	}

	http.SetCookie(w, cookie)
}

func (server *Server) setAccessTokenCookie(w http.ResponseWriter, accessToken string, accessTokenExpiry time.Time) {
	server.setCookie(w, "access_token", accessToken, "/", accessTokenExpiry)
}

func (server *Server) removeCookie(w http.ResponseWriter, name, path string) {
	cookie := http.Cookie{
		Name:    name,
		Value:   "",
		Path:    path,
		Expires: time.Unix(0, 0),
	}
	http.SetCookie(w, &cookie)
}
func (server *Server) addUserCookies(w http.ResponseWriter,
	accessToken,
	refreshToken string,
	accessTokenExpiry,
	refreshTokenExpiry time.Time,
) {
	server.setAccessTokenCookie(w, accessToken, accessTokenExpiry)
	server.setCookie(
		w,
		"refresh_token",
		refreshToken,
		"/v1/accouts/tokens/renew-access-token",
		refreshTokenExpiry)

}
func (server *Server) removeUserCookies(w http.ResponseWriter) {
	server.removeCookie(w, "access_token", "/")
	server.removeCookie(w, "refresh_token", "/v1/accouts/tokens/renew-access-token")
}
