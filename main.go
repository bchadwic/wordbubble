package main

import (
	"errors"
	"net/http"
	"os"

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

func main() {
	logger := newLogger("main")
	timer := util.NewTime()
	util.SigningKey = func() []byte {
	return []byte(os.Getenv("WB_SIGNING_KEY"))
	}
	wbRepo := wb.NewWordBubbleRepo(newLogger("wb_repo"))
	usersRepo := user.NewUserRepo(newLogger("users_repo"))
	authRepo := auth.NewAuthRepo(newLogger("auth_repo"))

	app := app.NewApp(
		auth.NewAuthService(newLogger("auth"), authRepo, timer, os.Getenv("WB_SIGNING_KEY")),
		user.NewUserService(newLogger("users"), usersRepo),
		wb.NewWordBubblesService(newLogger("wordbubbles"), wbRepo),
		newLogger("app"),
		timer,
	)

	http.HandleFunc("/signup", app.Signup)
	http.HandleFunc("/login", app.Login)
	http.HandleFunc("/token", app.Token)
	http.HandleFunc("/push", app.Push)
	http.HandleFunc("/pop", app.Pop)

	logger.Info("starting refresh token cleaner with an interval of: %gs", auth.RefreshTokenCleanerRate.Seconds())
	app.BackgroundCleaner(authRepo) // TODO make a timer interface to pass in

	logger.Info("starting server on port %s", port)
	err := http.ListenAndServe(port, nil)
	if errors.Is(err, http.ErrServerClosed) {
		logger.Info("server closed")
	} else if err != nil {
		logger.Error("could not start server %s", err)
		os.Exit(1)
	}
}
