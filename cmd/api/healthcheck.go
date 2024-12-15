package main

import (
	"net/http"
)

// Declare a handler which writes a plain text response with information about the application status, operating environment and version
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an envelope map containing the data for the response. Notice that the way we've constructed this means the environment and version data will now be nested
	// under a system_info key in the JSON response
	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}

	// Use the json.NewEncoder() function to initialize a json.Encoder instance that writes to the http.ResponseWriter.
	// Then we call its Encode() method, passing in the data that we want to encode to JSON (which in this case is the map above)
	// If the data can be successfully encoded to JSON, it will then be written to our http.ResponseWriter
	//err := json.NewEncoder(w).Encode(data)
	//if err != nil {
	//app.logger.Println(err)
	//http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	//}
	// This might work, however, when we call json.NewEncoder(w).Encode(data) the JSON is created and written to the http.ResponseWriter in a single step
	// Which means there's no opportunity to set HTTP response headers conditionally based on whether the Encode() method returns an error or not
	// Imagine, for example, that you want to set a Cache-Control header on a successful response, but not set a Cache-Control header if the JSON encoding fails
	// And you have to return an error response
	// Implementing that cleanly while using the json.Encoder pattern is quite difficult.
	// You could set the Cache-Control header and then delete it from the header map again in the event of an error - but that's pretty hacky.
}
