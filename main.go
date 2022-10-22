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

// @title                      wordbubble REST API
// @version                    1.0
// @description                wordbubble REST API interacts with auth and wordbubble data
// @contact.name               Ben Chadwick
// @contact.url                https://github.com/bchadwic
// @contact.email              benchadwick87@gmail.com
// @license.name               Apache 2.0
// @license.url                http://www.apache.org/licenses/LICENSE-2.0.html
// @host                       https://api.wordbubble.com
// @BasePath                   /v1
// @securityDefinitions.apikey ApiKeyAuth
// @tokenUrl                   https://api.wordbubble.com/token
// @in                         header
// @name                       Authorization
// @description                JWT access token retrieved from using a refresh token, gathered from /signup, /login, or /token
func run(cfg cfg.Config) error {
	logger := cfg.NewLogger("run")

	logger.Info("initializing repos and services")
	authRepo := auth.NewAuthRepo(cfg)
	usersRepo := user.NewUserRepo(cfg)
	wbRepo := wb.NewWordbubbleRepo(cfg)

	authService := auth.NewAuthService(cfg, authRepo)
	userService := user.NewUserService(cfg, usersRepo)
	wbService := wb.NewWordbubblesService(cfg, wbRepo)

	logger.Info("creating app")
	app := app.NewApp(cfg, authService, userService, wbService)

	logger.Info("attaching routes to app")
	http.HandleFunc("/v1/signup", app.Signup)
	http.HandleFunc("/v1/login", app.Login)
	http.HandleFunc("/v1/token", app.Token)
	http.HandleFunc("/v1/push", app.Push)
	http.HandleFunc("/v1/pop", app.Pop)

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
