package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"islamghany.greenlight/internals/data"
	"islamghany.greenlight/internals/marshing"
	"islamghany.greenlight/internals/validator"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {

	data := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}
	err := app.writeJson(w, http.StatusOK, data, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}

	v := validator.New()

	q := r.URL.Query()

	input.Title = app.readString(q, "title", "")
	input.Genres = app.readCSV(q, "genres", []string{})
	input.Filters.Page = app.readInt(q, "page", 1, v)
	input.Filters.PageSize = app.readInt(q, "page_size", 20, v)
	input.Filters.Sort = app.readString(q, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}
	input.Filters.ValidateFilters(v)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	movies, metadata, err := app.models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJson(w, http.StatusOK, envelope{"metadata": metadata, "movies": movies}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

const MAX_UPLOAD_SIZE = 1024 * 1024 * 3 // 3MB

// res, err := app.cld.Upload.Destroy(context.Background(), uploader.DestroyParams{
// 	PublicID: "movies/dkwpwnbbamephv0li1x5",
// })

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {

	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)
	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		app.badRequestResponse(w, r, errors.New("The uploaded file is too big. Please choose an file that's less than 1MB in size"))
		return
	}

	// The argument to FormFile must match the name attribute
	// of the file input on the frontend
	file, _, err := r.FormFile("image")
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	defer file.Close()

	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	filetype := http.DetectContentType(buff)
	if filetype != "image/jpeg" && filetype != "image/png" && filetype != "images/jpg" {
		app.badRequestResponse(w, r, errors.New("The provided file format is not allowed. Please upload a JPEG or PNG image"))
		return
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern

	fileType := strings.Split(filetype, "/")[1]
	tempFile, err := ioutil.TempFile("uploads", "upload-*."+fileType)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// write this byte array to our temporary file
	tempFile.Write(fileBytes)
	// We can choose to have these files deleted on program close
	defer os.Remove(tempFile.Name())
	//r.ParseMultipartForm(0)
	//input := r.MultipartForm.Value

	year, err := strconv.ParseInt(r.FormValue("year"), 10, 64)

	if err != nil {
		app.badRequestResponse(w, r, errors.New("runtime must be a number"))
		return
	}
	runtime, err := strconv.ParseInt(r.FormValue("runtime"), 10, 64)

	if err != nil {
		app.badRequestResponse(w, r, errors.New("runtime must be a number"))
		return
	}
	movie := &data.Movie{
		Title:   r.FormValue("title"),
		Year:    int32(year),
		Runtime: data.Runtime(runtime),
		Genres:  strings.Split(r.FormValue("genres"), ","),
	}

	// Initialize a new Validator.
	v := validator.New()

	// Call the ValidateMovie() function and return a response containing the errors if
	// any of the checks fail.
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	res, err := app.uploadImage(tempFile.Name(), uploader.UploadParams{
		Folder: "movies",
	})

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	movie.ImageID = res.PublicID
	movie.ImageURL = res.SecureURL

	user := app.contextGetUser(r)

	err = app.models.Movies.Insert(movie, user.ID)
	if err != nil {
		err = app.destroyImage(movie.ImageID)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	err = app.writeJson(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.Movies.CacheGetMovie(data.MoviesKey(id))

	if err != nil {
		movie, err = app.models.Movies.Get(id)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notFoundResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}
	}

	err = app.writeJson(w, http.StatusOK, envelope{"movie": movie}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.Movies.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		movie.Title = *input.Title
	}
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	// Validate the updated movie record, sending the client a 422 Unprocessable Entity
	// response if any checks fail.
	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Write the updated movie record in a JSON response.
	err = app.writeJson(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJson(w, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) GetMostViewsMovivesHandler(w http.ResponseWriter, r *http.Request) {

	res, err := app.models.Movies.CacheGetMostViews()
	if err != nil {
		movies, err := app.models.Movies.GetMostViews()
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		err = app.writeJson(w, http.StatusOK, envelope{"movies": movies}, nil)

		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		app.background(func() {
			data, err := marshing.MarshalBinary(envelope{"movies": movies})
			if err != nil {
				app.logger.PrintError(err, nil)
			}
			err = app.models.Movies.CacheSetMostViews(string(data))
			if err != nil {
				app.logger.PrintError(err, nil)
			}
		})
		return
	}
	app.writeJsonString(w, *res)
}

func (app *application) GetMostLikedMovivesHandler(w http.ResponseWriter, r *http.Request) {
	res, err := app.models.Movies.CacheGetMostLikes()
	if err != nil {
		movies, err := app.models.Movies.GetMostLikes()
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		err = app.writeJson(w, http.StatusOK, envelope{"movies": movies}, nil)

		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		app.background(func() {
			data, err := marshing.MarshalBinary(envelope{"movies": movies})
			if err != nil {
				app.logger.PrintError(err, nil)
			}
			err = app.models.Movies.CacheSetMostLikes(string(data))
			if err != nil {
				app.logger.PrintError(err, nil)
			}
		})
		return
	}
	app.writeJsonString(w, *res)
}
