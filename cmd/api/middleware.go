package main

import (
	"fmt"
	"net/http"
)

// It's important to note that our middleware will only recover panics that happen in the same goroutine that executed it.
// If you spin aditional goroutines from within your handlers and there is a change to panic, handle those panics inside those goroutines.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event of a panic) as Go unwinds the stack.
		defer func() {
			// Use the builtin recover function to check if there has been a panic or not.
			if err := recover(); err != nil {
				// If there was a panic, set a "Connection: close" header on the response.
				// This acts as a trigger to make Go's HTTP server automatically close the current connection after a response has been sent.
				w.Header().Set("Connection", "close")
				// The value retuned by recover() has the type interface{}
				// fmt.Errorf() to normalize it into an error and call our serverErrorResponse() helper.
				// In turn, this will log the error using our custom Logger type at the ERROR level and send the client a 500
				// Internal server error response.
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
