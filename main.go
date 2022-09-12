package main

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/bchadwic/wordbubble/auth"
	"github.com/bchadwic/wordbubble/user"
	"github.com/bchadwic/wordbubble/util"
	"github.com/bchadwic/wordbubble/wb"
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

func main() {
	dataSource := wb.NewDataSource(newLogger("datasource"))
	authSource := auth.NewAuthSource(newLogger("authsource"))
	app := &App{
		newLogger("app"),
		auth.NewAuth(authSource, newLogger("auth"), os.Getenv("WB_SIGNING_KEY")),
		user.NewUsersService(dataSource, newLogger("users")),
		wb.NewWordBubblesService(dataSource, newLogger("wordbubbles")),
	}

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
