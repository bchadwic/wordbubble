package app

import (
	"net/http"

	cfg "github.com/bchadwic/wordbubble/internal/config"
	"github.com/bchadwic/wordbubble/internal/service/auth"
	"github.com/bchadwic/wordbubble/internal/service/user"
	"github.com/bchadwic/wordbubble/internal/service/wb"
	"github.com/bchadwic/wordbubble/model/resp"
	"github.com/bchadwic/wordbubble/util"
)

type app struct {
	auth        auth.AuthService
	users       user.UserService
	wordbubbles wb.WordbubbleService
	log         util.Logger
	timer       util.Timer
}

func NewApp(cfg cfg.Config, authService auth.AuthService, userService user.UserService, wbService wb.WordbubbleService) *app {
	return &app{
		auth:        authService,
		users:       userService,
		wordbubbles: wbService,
		log:         cfg.NewLogger("app"),
		timer:       cfg.Timer(),
	}
}

// TODO make this entire file better
func (wb *app) BackgroundCleaner(authCleaner auth.AuthCleaner) {
	const refreshTokenTimeLimit = 60
	go func() {
		for range wb.timer.Tick(auth.RefreshTokenCleanerRate) {
			_ = authCleaner.CleanupExpiredRefreshTokens(wb.timer.Now().Unix() - refreshTokenTimeLimit)
		}
	}()
}

func (wb *app) errorResponse(err error, w http.ResponseWriter) {
	switch t := err.(type) {
	case *resp.StatusNoContent:
		wb.log.Warn("%d - %s", t.Code, t.Error())
		w.WriteHeader(t.Code)
		w.Write([]byte(t.Message))
	case *resp.StatusBadRequest:
		wb.log.Warn("%d - %s", t.Code, t.Error())
		w.WriteHeader(t.Code)
		w.Write([]byte(t.Message))
	case *resp.StatusUnauthorized:
		wb.log.Warn("%d - %s", t.Code, t.Error())
		w.WriteHeader(t.Code)
		w.Write([]byte(t.Message))
	case *resp.StatusMethodNotAllowed:
		wb.log.Warn("%d - %s", t.Code, t.Error())
		w.WriteHeader(t.Code)
		w.Write([]byte(t.Message))
	case *resp.StatusConflict:
		wb.log.Warn("%d - %s", t.Code, t.Error())
		w.WriteHeader(t.Code)
		w.Write([]byte(t.Message))
	case *resp.StatusInternalServerError:
		wb.log.Warn("%d - %s", t.Code, t.Error())
		w.WriteHeader(t.Code)
		w.Write([]byte(t.Message))
	default:
		wb.log.Error("%d - %s", http.StatusInternalServerError, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(resp.Unknown)
	}
}
