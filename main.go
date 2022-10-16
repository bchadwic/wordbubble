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

// @title          Swagger Example API
// @version        1.0
// @description    This is a sample server celler server.
// @termsOfService http://swagger.io/terms/

// @contact.name  API Support
// @contact.url   http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url  http://www.apache.org/licenses/LICENSE-2.0.html

// @host     localhost:8080
// @BasePath /api/v1

// @securityDefinitions.basic BasicAuth
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
	http.HandleFunc("/pop/", app.Pop)

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
