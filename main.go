package main

import (
	"errors"
	"net/http"
	"os"
	"time"
)

func main() {
	port := ":8080"
	logger := NewLogger(os.Getenv("WB_LOG_LEVEL"))
	dataSource := NewDataSource()
	authSource := NewAuthSource()

	go func() {
		for range time.Tick(time.Minute * 30) {
			authSource.CleanupExpiredRefreshTokens(logger)
		}
	}()

	app := &App{
		logger,
		NewAuth(authSource, os.Getenv("WB_SIGNING_KEY")),
		NewUsersService(dataSource),
		NewWordBubblesService(dataSource),
	}

	http.HandleFunc("/register", app.Register)
	http.HandleFunc("/token", app.Token)
	http.HandleFunc("/refresh_token", app.RefreshToken)
	http.HandleFunc("/push", app.Push)
	http.HandleFunc("/pop", app.Pop)

	logger.Info("starting server on port %s", port)
	err := http.ListenAndServe(port, nil)
	if errors.Is(err, http.ErrServerClosed) {
		logger.Info("server closed")
	} else if err != nil {
		logger.Error("could not start server %s", err)
		os.Exit(1)
	}
}
