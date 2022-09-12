package main

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/bchadwic/wordbubble/app"
	"github.com/bchadwic/wordbubble/internal/auth"
	"github.com/bchadwic/wordbubble/internal/user"
	"github.com/bchadwic/wordbubble/internal/wb"
	"github.com/bchadwic/wordbubble/util"
)

var newLogger = func(namespace string) util.Logger {
	return util.NewLogger(namespace, os.Getenv("WB_LOG_LEVEL"))
}

var port = func() string {
	if p := os.Getenv("WB_PORT"); p != "" {
		return p
	}
	return ":8080"
}()

// MAKE AN INTERNAL PACKAGE FOR ALL OF YOUR SERVICE AND REPO LAYER
func main() {
	dataSource := wb.NewDataSource(newLogger("datasource"))
	authSource := auth.NewAuthSource(newLogger("authsource"))
	app := app.NewApp(
		auth.NewAuth(authSource, newLogger("auth"), os.Getenv("WB_SIGNING_KEY")),
		user.NewUsersService(dataSource, newLogger("users")),
		wb.NewWordBubblesService(dataSource, newLogger("wordbubbles")),
		newLogger("app"),
	)

	http.HandleFunc("/signup", app.Signup)
	http.HandleFunc("/login", app.Login)
	http.HandleFunc("/token", app.Token)
	http.HandleFunc("/push", app.Push)
	http.HandleFunc("/pop", app.Pop)

	logger := newLogger("main")
	logger.Info("starting server on port %s", port)
	go func() {
		for range time.Tick(auth.RefreshTokenCleanerRate) {
			authSource.CleanupExpiredRefreshTokens()
		}
	}()
	err := http.ListenAndServe(port, nil)
	if errors.Is(err, http.ErrServerClosed) {
		logger.Info("server closed")
	} else if err != nil {
		logger.Error("could not start server %s", err)
		os.Exit(1)
	}
}
