package app

import (
	"net/http"

	cfg "github.com/bchadwic/wordbubble/internal/config"
	"github.com/bchadwic/wordbubble/internal/service/auth"
	"github.com/bchadwic/wordbubble/internal/service/user"
	"github.com/bchadwic/wordbubble/internal/service/wb"
	"github.com/bchadwic/wordbubble/resp"
	"github.com/bchadwic/wordbubble/util"
)

type app struct {
	auth        auth.AuthService
	users       user.UserService
	wordbubbles wb.WordBubbleService
	log         util.Logger
	timer       util.Timer
}

func NewApp(cfg cfg.Config, authService auth.AuthService, userService user.UserService, wbService wb.WordBubbleService) *app {
	return &app{
		auth:        authService,
		users:       userService,
		wordbubbles: wbService,
		log:         cfg.NewLogger("app"),
		timer:       cfg.Timer(),
	}
}

// TODO make this better
func (wb *app) BackgroundCleaner(authCleaner auth.AuthCleaner) {
	const refreshTokenTimeLimit = 60
	go func() {
		for range wb.timer.Tick(auth.RefreshTokenCleanerRate) {
			_ = authCleaner.CleanupExpiredRefreshTokens(wb.timer.Now().Unix() - refreshTokenTimeLimit)
		}
	}()
}

func (wb *app) errorResponse(err error, w http.ResponseWriter) {
	wb.log.Error("an error occurred %s", err.Error())
	switch t := err.(type) {
	case *resp.ErrorResponse:
		w.WriteHeader(t.Code)
		w.Write(t.Message)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(resp.Unknown)
	}
}
