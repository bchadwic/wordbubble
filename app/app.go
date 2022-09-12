package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bchadwic/wordbubble/internal/auth"
	"github.com/bchadwic/wordbubble/internal/user"
	"github.com/bchadwic/wordbubble/internal/wb"
	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/util"
)

type App struct {
	auth auth.AuthService
	users user.UserService
	wordbubbles   wb.WordBubbleService
	log         util.Logger
}

func NewApp(authService auth.AuthService, userService user.UserService, wbService wb.WordBubbleService, log util.Logger) *App {
	return &App{authService, userService, wbService, log}
}

func (app *App) BackgroundCleaner(authCleaner auth.AuthCleaner) {
	go func() {
		for range time.Tick(auth.RefreshTokenCleanerRate) {
			authCleaner.CleanupExpiredRefreshTokens()
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

	accessToken, err := app.auth.GenerateAccessToken(user.Id)
	if err != nil {
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
	}{refreshToken, accessToken}
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
		user     string `json:"user"`
		password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		app.log.Error("could not decode user from body: %s", err)
		app.respond("could not parse a user from the request body", http.StatusBadRequest, w)
		return
	}

	AuthenticateUser := app.users.RetrieveAuthenticatedUserByString(reqBody.user, reqBody.password)
	if AuthenticateUser == nil {
		app.respond("could not authenticate user using credentials passed", http.StatusUnauthorized, w)
		return
	}

	accessToken, err := app.auth.GenerateAccessToken(AuthenticateUser.Id)
	if err != nil {
		app.respond(err.Error(), http.StatusInternalServerError, w)
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
	}{refreshToken, accessToken}
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
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		app.log.Error("could not decode refresh token from body, error: %s", err)
		app.respond("could not parse refresh token from the request body", http.StatusBadRequest, w)
		return
	}

	userId, err := app.auth.GetUserIdFromTokenString(reqBody.RefreshToken)
	if err != nil {
		app.respond(err.Error(), http.StatusUnauthorized, w)
		return
	}

	timeBeforeExpiration, err := app.auth.VerifyTokenAgainstAuthSource(userId, reqBody.RefreshToken)
	if err != nil {
		app.respond(err.Error(), http.StatusUnauthorized, w)
		return
	}

	fmt.Printf("timeBeforeExpiration: %d, ImminentExpirationWindow: %d\n", timeBeforeExpiration, auth.ImminentExpirationWindow)
	var latestRefreshToken string
	if timeBeforeExpiration < auth.ImminentExpirationWindow {
		latestRefreshToken = app.auth.GetOrCreateLatestRefreshToken(userId)
	}

	accessToken, err := app.auth.GenerateAccessToken(userId)
	if err != nil {
		app.respond(err.Error(), http.StatusInternalServerError, w)
		return
	}

	resp := struct {
		RefreshToken string `json:"refresh_token,omitempty"`
		AccessToken  string `json:"access_token"`
	}{latestRefreshToken, accessToken}
	app.log.Info("generated token was successful, sending back token response")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (app *App) Push(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.log.Error("invalid http method: %s", r.Method)
		app.respond("invalid http method", http.StatusMethodNotAllowed, w)
		return
	}

	authValue := r.Header.Get("authorization")
	if authValue == "" {
		app.log.Error("authorization was not passed")
		app.respond("authorization header is required for pushing a wordbubble", http.StatusUnauthorized, w)
		return
	}

	splitToken := strings.Split(authValue, "Bearer ")
	if len(splitToken) < 2 {
		app.log.Error("authorization value didn't specifiy token type as bearer")
		app.respond("a bearer token is required for pushing a wordbubble", http.StatusUnauthorized, w)
		return
	}

	token := splitToken[1] // grab the token from the Bearer string
	userId, err := app.auth.GetUserIdFromTokenString(token)
	if err != nil {
		app.respond(err.Error(), http.StatusUnauthorized, w)
		return
	}

	var wb model.WordBubble // finally we are authenticated! Let's insert a wordbubble
	if err = json.NewDecoder(r.Body).Decode(&wb); err != nil {
		app.log.Error("could not decode wordbubble from body: %s", err)
		app.respond("could not parse a wordbubble from request body", http.StatusBadRequest, w)
		return
	}

	if err = app.wordbubbles.ValidWordBubble(&wb); err != nil {
		app.respond(err.Error(), http.StatusBadRequest, w)
		return
	}

	if err = app.wordbubbles.UserHasAvailability(userId); err != nil {
		app.respond(err.Error(), http.StatusConflict, w)
		return
	}

	if err = app.wordbubbles.AddNewWordBubble(userId, &wb); err != nil {
		app.respond(err.Error(), http.StatusInternalServerError, w)
		return
	}

	if wb.Text == "teapot" {
		app.log.Info("found ourselves a teapot")
		app.respond("here is some tea for you", http.StatusTeapot, w)
		return
	}

	app.log.Info("successfully created a wordbubble for %d", userId)
	app.respond("thank you!", http.StatusCreated, w)
}

func (app *App) Pop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		app.log.Error("invalid http method: %s", r.Method)
		app.respond("invalid http method", http.StatusMethodNotAllowed, w)
		return
	}

	var reqBody struct {
		UserStr string `json:"user"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		app.log.Error("could not decode identity from body: %s", err)
		app.respond("could not parse a user from request body", http.StatusBadRequest, w)
		return
	}

	user := app.users.RetrieveUserByString(reqBody.UserStr)
	if user == nil {
		app.respond(fmt.Sprintf("could not resolve user `%s`", reqBody.UserStr), http.StatusBadRequest, w)
		return
	}

	wordbubble := app.wordbubbles.RemoveAndReturnLatestWordBubbleForUserId(user.Id)
	if wordbubble == nil {
		app.respond(fmt.Sprintf("an internal error occurred while fetching wordbubble for %s", reqBody.UserStr), http.StatusInternalServerError, w)
		return
	}

	if wordbubble.Text == "" {
		app.respond(fmt.Sprintf("no wordbubble found for %s", reqBody.UserStr), http.StatusNoContent, w)
		return
	}

	app.log.Info("successfully popped a wordbubble for %d", user.Id)
	app.respond(wordbubble.Text, http.StatusOK, w)
}

func (app *App) respond(response string, statusCode int, w http.ResponseWriter) {
	w.WriteHeader(statusCode)
	w.Write([]byte(response + "\n")) // temporary, soon to be a struct
}
