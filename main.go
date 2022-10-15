package main

import (
	"errors"
	"net/http"
	"os"

	"github.com/bchadwic/wordbubble/app"
	cfg "github.com/bchadwic/wordbubble/internal/config"
	"github.com/bchadwic/wordbubble/internal/service/auth"
	"github.com/bchadwic/wordbubble/internal/service/user"
	"github.com/bchadwic/wordbubble/internal/service/wb"
)

func main() {
	cfg := cfg.NewConfig()
	if cfg == nil || run(cfg) != nil {
		os.Exit(1)
	}
}

func run(cfg cfg.Config) error {
	logger := cfg.NewLogger("run")

	logger.Info("initializing repos and services")
	authRepo := auth.NewAuthRepo(cfg)
	usersRepo := user.NewUserRepo(cfg)
	wbRepo := wb.NewWordBubbleRepo(cfg)

	authService := auth.NewAuthService(cfg, authRepo)
	userService := user.NewUserService(cfg, usersRepo)
	wbService := wb.NewWordBubblesService(cfg, wbRepo)

	logger.Info("creating app")
	app := app.NewApp(cfg, authService, userService, wbService)

	logger.Info("attaching routes to app")
	http.HandleFunc("/signup", app.Signup)
	http.HandleFunc("/login", app.Login)
	http.HandleFunc("/token", app.Token)
	http.HandleFunc("/push", app.Push)
	http.HandleFunc("/pop", app.Pop)

	logger.Info("starting refresh token cleaner with an interval of: %gs", auth.RefreshTokenCleanerRate.Seconds())
	app.BackgroundCleaner(authRepo)

	logger.Info("starting server on port %s", cfg.Port())
	err := http.ListenAndServe(cfg.Port(), nil)
	if errors.Is(err, http.ErrServerClosed) {
		logger.Info("server closed")
		return nil
	} else if err != nil {
		logger.Error("could not start server %s", err)
	}
	return err
}
