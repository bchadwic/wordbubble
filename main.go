package main

import (
	"errors"
	"net/http"
	"os"
)

func main() {
	port := ":8080"

	app := &App{
		NewLogger(os.Getenv("WB_LOG_LEVEL")),
		NewAuth(os.Getenv("WB_SIGNING_KEY")),
		NewUsersService(),
	}

	http.HandleFunc("/register", app.Register)
	http.HandleFunc("/token", app.Token)
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
