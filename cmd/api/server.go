package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	// Declare a HTTP server using the same settings as in our main() function
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Create a shutdownError channel.
	// We will use this to receive any errors returned
	// by the graceful Shutdown() function
	shutdownError := make(chan error)

	// Start a background goroutine
	go func() {
		// Create a quit channel which carries os.Signal values.
		// Intercept the signals
		quit := make(chan os.Signal, 1)

		// Use signal.Notify() to listen for incoming SIGINT and SIGTERM signals and
		// relay them to the quit channel. Any other signals will not be caight by
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// Read the signal from the quit channel. This code will block until a signal is received
		s := <-quit

		// Update the log entry to say "shutting down server" instead of "caught signal".
		app.logger.PrintInfo("shutting down server", map[string]string{
			"signal": s.String(),
		})

		// Create a context with a 5-second timeout.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Call Shutdown() on our server, passing in the context we just made.
		// Shutdown() will return nil if the graceful shutdown was successful
		// Or an error because the shutdown didn't complete before the 5 - second context
		// We relay this return value to the shutdownError channel
		shutdownError <- srv.Shutdown(ctx)
	}()

	// Likewise log a "starting server" message
	// Start the server as normal
	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})

	// Calling Shutdown() on our server will cause ListenAndServe() to immediately
	// return a http.ErrServerClosed error. So if we see this error, it is actually a good thing and an indication that the graceful shutdown has started.
	// So we check specifically for this, only returning the error if it is not http.ErrServerClosed.
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Otherwise, we wait to receive the return value from Shutdown() on the shutdownError channel.
	// If return value is an error, we know that there was a problem with the graceful shutdown and we return the error.
	err = <-shutdownError
	if err != nil {
		return err
	}

	// At this point we know that the graceful shutdown completed successfully and we log a "stopped server" message.
	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}
