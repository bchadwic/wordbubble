package app

import (
	"encoding/json"
	"net/http"

	"github.com/bchadwic/wordbubble/internal/auth"
	"github.com/bchadwic/wordbubble/internal/user"
	"github.com/bchadwic/wordbubble/internal/wb"
	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/resp"
	"github.com/bchadwic/wordbubble/util"
)

type App struct {
	auth        auth.AuthService
	users       user.UserService
	wordbubbles wb.WordBubbleService
	log         util.Logger
	timer       util.Timer
}

func NewApp(authService auth.AuthService, userService user.UserService, wbService wb.WordBubbleService, log util.Logger, timer util.Timer) *App {
	return &App{
		auth:        authService,
		users:       userService,
		wordbubbles: wbService,
		log:         log,
		timer:       timer,
	}
}

const refreshTokenTimeLimit = 60

func (app *App) BackgroundCleaner(authCleaner auth.AuthCleaner) {
	go func() {
		for range app.timer.Tick(auth.RefreshTokenCleanerRate) {
			_ = authCleaner.CleanupExpiredRefreshTokens(app.timer.Now().Unix() - refreshTokenTimeLimit)
		}
	}()
}

func (app *App) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.log.Error("invalid http method: %s", r.Method)
		app.respond("invalid http method", http.StatusMethodNotAllowed, w)
		return
	}

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		app.log.Error("could not decode user from body: %s", err)
		app.respond("could not parse a user from the request body", http.StatusBadRequest, w)
		return
	}

	if err := util.ValidUser(&user); err != nil {
		app.respond(err.Error(), http.StatusBadRequest, w)
		return
	}

	if err := app.users.AddUser(&user); err != nil {
		app.respond(err.Error(), http.StatusInternalServerError, w)
		return
	}

	refreshToken, err := app.auth.GenerateRefreshToken(user.Id)
	if err != nil {
		app.respond(err.Error(), http.StatusInternalServerError, w)
		return
	}

	resp := struct {
		RefreshToken string `json:"refresh_token"`
		AccessToken  string `json:"access_token"`
	}{refreshToken, app.auth.GenerateAccessToken(user.Id)}
	app.log.Info("generated token was successful, sending back token response")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (app *App) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.log.Error("invalid http method: %s", r.Method)
		app.respond("invalid http method", http.StatusMethodNotAllowed, w)
		return
	}

	var reqBody struct {
		User     string `json:"user"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		app.log.Error("could not decode user from body: %s", err)
		app.respond("could not parse a user from the request body", http.StatusBadRequest, w)
		return
	}

	AuthenticateUser := app.users.RetrieveAuthenticatedUserByString(reqBody.User, reqBody.Password)
	if AuthenticateUser == nil {
		app.respond("could not authenticate user using credentials passed", http.StatusUnauthorized, w)
		return
	}

	refreshToken, err := app.auth.GenerateRefreshToken(AuthenticateUser.Id)
	if err != nil {
		app.respond(err.Error(), http.StatusInternalServerError, w)
		return
	}

	resp := struct {
		RefreshToken string `json:"refresh_token"`
		AccessToken  string `json:"access_token"`
	}{refreshToken, app.auth.GenerateAccessToken(AuthenticateUser.Id)}
	app.log.Info("generated token was successful, sending back token response")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (app *App) Token(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.log.Error("invalid http method: %s", r.Method)
		app.respond("invalid http method", http.StatusMethodNotAllowed, w)
		return
	}

	var reqBody struct {
		TokenString string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		app.log.Error("could not decode refresh token from body, error: %s", err)
		app.respond("could not decode the request body", http.StatusBadRequest, w)
		return
	}

	token, err := auth.RefreshTokenFromTokenString(reqBody.TokenString)
	if err != nil {
		app.log.Error("could not generate a token struct from token string passed in: %s", err)
		app.respond("could not parse refresh token from the request body", http.StatusBadRequest, w)
		return
	}
	if err = app.auth.ValidateRefreshToken(token); err != nil {
		app.respond(err.Error(), http.StatusUnauthorized, w)
		return
	}

	var latestRefreshToken string
	if token.IsNearEndOfLife() {
		latestRefreshToken, _ = app.auth.GenerateRefreshToken(token.UserId())
	}

	resp := struct {
		RefreshToken string `json:"refresh_token,omitempty"`
		AccessToken  string `json:"access_token"`
	}{latestRefreshToken, app.auth.GenerateAccessToken(token.UserId())}
	app.log.Info("generated token was successful, sending back token response")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (app *App) respond(response string, statusCode int, w http.ResponseWriter) {
	w.WriteHeader(statusCode)
	w.Write([]byte(response)) // temporary, soon to be a struct
}

func writeResponse(w http.ResponseWriter, statusCode int, resp any) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(resp)
}

func (app *App) errorResponse(err error, w http.ResponseWriter) {
	app.log.Error("app - an error occurred %w", err)
	switch t := err.(type) {
	case *resp.ErrorResponse:
		w.WriteHeader(t.Code)
		w.Write(t.Message)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(resp.Unknown)
	}
}
