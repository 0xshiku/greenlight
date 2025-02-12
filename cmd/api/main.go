package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"greenlight/internal/data"
	"greenlight/internal/jsonlog"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// Declare a string containing the application version number. Later in the book we'll generate
// this automatically at build time, but for now we'll just store the version number as a hard-coded global constant
const version = "1.0.0"

// Define a config struct to hold all the configuration settings for our application.
// For now, the only configuration settings will be the network port that we want the server to listen on.
// And the name of the current operating environment for the application (development, staging, production, etc...)
// We will read in these configuration settings from command-line flags when the application starts.
// Add a db struct field to hold the configuration settings for our database connection pool.
// For now this only holds the DSN, which we will read in from a command-line flag.
// Add maxOpenConns, maxIdleConns and maxIdleTime fields to hold the configuration settings for the connection pool
type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
		sdsd         string
	}
}

// Define an application struct to hold the dependencies for our HTTP handlers, helpers, and middleware.
// At the moment this only contains a copy of the config struct and a logger, but it will grow to include a lot more as our build progresses.
// Add a models field to hour new Models struct
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
}

func main() {
	// Declare an instance of the config struct
	var cfg config

	// Read the value of the port and env command-line flags into the config struct.
	// We default to using the port number 4000 and the environment "development" if no corresponding flags are provided.
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	// Use the value of the GREENLIGHT_DB_DSN environment variable as the default value for our db-dsn command-line flag.
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")
	// Read the connection pool settings from command-line flags into the config struct.
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.Parse()

	// Initialize a new logger which writes messages to the standard out stream, prefixed with the current date and time.
	// Initialize a new jsonlog.Logger which writes any messages *at or above* the INFO
	// severity level to the standard out stream.
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// Call the openDB() helper function (see below) to create the connection pool, passing in the config struct.
	// If this returns an error, we log it and exit the application immediately.
	db, err := openDB(cfg)
	if err != nil {
		// Use the PrintFatal() method to write a log entry containing the error at the FATAL level and exit.
		// We have no additional properties to include in the log entry.
		// So we pass nil as the second parameter.
		logger.PrintFatal(err, nil)
	}

	// Defer a call to db.Close() so that the connection pool is closed before the main() function exists
	defer db.Close()
	// Likewise use the PrintInfo() method to write a message at the INFO level.
	logger.PrintInfo("database connection pool established", nil)

	// Declare an instance of the application struct, containing the config struct and the logger
	// Use the data.NewModels() function to initialize a Models struct, passing in the connection pool as a parameter
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	// Declare an HTTP server with some sensible timeout settings, which listens on the port provided in the config struct and uses the serve mux we created above as the handler
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Again, we use the PrintInfo method to write a starting server message at the INFO level.
	// But this time we pass a map containing additional properties
	// The operating environment and server address as the final parameter.
	logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  cfg.env,
	})

	// Because the err variable is now already declared in the code above, we need to use the = operator here, instead of the := operator.
	err = srv.ListenAndServe()
	// Use the PrintFatal() method to log the error and exit.
	logger.PrintFatal(err, nil)
}

// The openDB() function returns a sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {
	// Use sql.Open() to create an empty connection pool, using the DSN from the config struct.
	// How does the sql.DB connection pool work?
	// Important thing to know is, sql.DB pool contains two types of connections, 'in-use' and 'idle' connections.
	// A connection is marked as in-use when you are using it to perform a database task, such as executing a SQL statement.
	// When the task is complete the connection is then marked as idle.
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Set the maximum number of open (in-use +idle) connections in the pool. Note that passing a value less than or equal to 0 will mean there is no limit.
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	// Set the maximum number of idle connections in the pool. Again, passing a value less than or equal to 0 mean there is no limit
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	// Use the time.ParseDuration() function to convert the idle timeout durantion string to a time.Duration type.
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	// Set the maximum idle timeout
	db.SetConnMaxIdleTime(duration)

	// Create a context with a 5 seconds timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use PingContext() to establish a new connection to the database, passing in the context
	// We created above as a parameter. If the connection couldn't be established successfully withing the 5 seconds deadline.
	// Then this will return an error.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	// Return the sql.DB connection pool.
	return db, nil
}
