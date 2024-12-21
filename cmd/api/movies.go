package main

import (
	"errors"
	"fmt"
	"greenlight/internal/data"
	"greenlight/internal/validator"
	"net/http"
)

// Add a shownMovieHandler for the "GET /v1/movies/:id" endpoint.
// For now, we retrieve the interpolated  "id" parameter from the current URL and include it in a placeholder response
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	// We can them use the ByName() method to get the value of the "id" parameter from the slice.
	// In our project all movies will have a unique positive integer ID, but the value returned by ByName() is always a string.
	// So we try to convert it to a or is less than 1, we know the ID is invalid so we use the http.NotFound()
	// function to return a 404 Not Found response.
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	// Call the Get() method to fetch the data for a specific movie.
	// We also need to use the errors.Is() function, to check if it returns a data.ErrRecordNotFound error,
	// In which case we send a 404 Not Found response to the client.
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

	// Encode the struct to JSON and send it as the HTTP response
	// Create an envelope{"movie": movie} instance and pass it to writeJSON(), instead of passing the plain movie struct
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous struct to hold the information that we expect to be in the
	// HTTP request body (note that the field names and types in the struct are a subset of the Movie struct that we created earlier).
	// This struct will be our *target decode destination*
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	// Initialize a new json.Decoder instance which reads from the request body, and then the Decode() method to decode the body contents into the input struct.
	// Importantly, notice that when we call Decode() we pass a *pointer* to the input struct as the target decode destination. If there was an error during decoding,
	// we also use our generic errorResponse() helper to send the client a 400 Bad Request response containing the error message
	// We could use unmarshal, but it's more verbose and requires about 80% more memory.
	// Now use the new readJSON() helper to decode the request body into the input struct.
	// If this returns an error we send the client the error message along with a 400
	// Bad Request status, just like before
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy the values from the input struct to a new movie struct
	// The problem with decoding directly into a Movie struct is that a client could provide the keys id and version in their JSON request
	// and the corresponding values would be decoded without any error into the ID and Version fields of the Movie struct
	// Note that the movie variable contains a *pointer* to a Movie struct.
	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	// Initialize a new Validator instance
	v := validator.New()

	// Use the Valid() method to see if any of the checks failed. If they did, then use the failedValidationResponse() helper to send a response to the client, passing in the v.Errors map.
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the Insert() method on our movies model, passing in a pointer to the validated movie struct.
	// This will create a record in the database and update the movie struct with the system generated information.
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// When sending a HTTP response, we want to include a Location header to let the client know which URL they can find the newly created resource at.
	// We make an empty http.Header map and then use the Set() method to add a new Location header, interpolating the system-generated ID for our new movie in the URL
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	// Write a JSON response with a 201 Created status code, the movie data in the response body, and the Location header.
	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
