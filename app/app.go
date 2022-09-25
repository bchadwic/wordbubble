package app

import (
	"net/http"

	"github.com/bchadwic/wordbubble/internal/auth"
	"github.com/bchadwic/wordbubble/internal/user"
	"github.com/bchadwic/wordbubble/internal/wb"
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

func NewApp(authService auth.AuthService, userService user.UserService, wbService wb.WordBubbleService, log util.Logger, timer util.Timer) *app {
	return &app{
		auth:        authService,
		users:       userService,
		wordbubbles: wbService,
		log:         log,
		timer:       timer,
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
	wb.log.Error("an error occurred %w", err)
	switch t := err.(type) {
	case *resp.ErrorResponse:
		w.WriteHeader(t.Code)
		w.Write(t.Message)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(resp.Unknown)
	}
}
