package main

import (
	"errors"
	"net/http"
	"os"
)

func main() {
	port := ":8080"

	app := &App{
		logger: NewLogger(os.Getenv("WB_LOG_LEVEL")),
		users:  NewUsersService(os.Getenv("WB_SIGNING_KEY")),
	}

	http.HandleFunc("/signup", app.Signup)
	http.HandleFunc("/login", app.Login)
	http.HandleFunc("/push", app.Push)

	app.logger.Info("starting server on port %s", port)
	err := http.ListenAndServe(port, nil)
	if errors.Is(err, http.ErrServerClosed) {
		app.logger.Info("server closed")
	} else if err != nil {
		app.logger.Error("could not start server %s", err)
		os.Exit(1)
	}
}
