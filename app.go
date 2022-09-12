package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type App struct {
	logger Logger
	auth   Auth
	users  Users
	wbs    WordBubbles
}

func (app *App) respond(response string, statusCode int, w http.ResponseWriter) {
	w.WriteHeader(statusCode)
	w.Write([]byte(response + "\n")) // temporary, soon to be a struct
}

func (app *App) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.logger.Error("invalid http method: %s", r.Method)
		app.respond("invalid http method", http.StatusMethodNotAllowed, w)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		app.logger.Error("could not decode user from body: %s", err)
		app.respond("could not parse a user from the request body", http.StatusBadRequest, w)
		return
	}

	if err := ValidPassword(user.Password); err != nil {
		app.respond(err.Error(), http.StatusBadRequest, w)
		return
	}

	user.Username = strings.Trim(user.Username, " ")
	if err := app.users.ValidUser(&user); err != nil {
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
	app.logger.Info("generated token was successful, sending back token response")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (app *App) Login(w http.ResponseWriter, r *http.Request) {
	logger := app.logger
	logger.Info("handling request")

	if r.Method != http.MethodPost {
		logger.Error("invalid http method: %s", r.Method)
		app.respond("invalid http method", http.StatusMethodNotAllowed, w)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		logger.Error("could not decode user from body: %s", err)
		app.respond("could not parse a user from the request body", http.StatusBadRequest, w)
		return
	}

	user.Username = strings.Trim(user.Username, " ")
	if err := app.users.AuthenticateUser(&user); err != nil {
		app.respond(err.Error(), http.StatusUnauthorized, w)
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
	logger.Info("generated token was successful, sending back token response")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (app *App) Token(w http.ResponseWriter, r *http.Request) {
	logger := app.logger
	logger.Info("handling request")

	if r.Method != http.MethodPost {
		logger.Error("invalid http method: %s", r.Method)
		app.respond("invalid http method", http.StatusMethodNotAllowed, w)
		return
	}

	var reqBody struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		logger.Error("could not decode refresh token from body, error: %s", err)
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

	fmt.Printf("timeBeforeExpiration: %d, ImminentExpirationWindow: %d\n", timeBeforeExpiration, ImminentExpirationWindow)
	var latestRefreshToken string
	if timeBeforeExpiration < ImminentExpirationWindow {
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
	logger.Info("generated token was successful, sending back token response")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (app *App) Push(w http.ResponseWriter, r *http.Request) {
	logger := app.logger
	logger.Info("handling request")

	if r.Method != http.MethodPost {
		logger.Error("invalid http method: %s", r.Method)
		app.respond("invalid http method", http.StatusMethodNotAllowed, w)
		return
	}

	authValue := r.Header.Get("authorization")
	if authValue == "" {
		logger.Error("authorization was not passed")
		app.respond("authorization header is required for pushing a wordbubble", http.StatusUnauthorized, w)
		return
	}

	splitToken := strings.Split(authValue, "Bearer ")
	if len(splitToken) < 2 {
		logger.Error("authorization value didn't specifiy token type as bearer")
		app.respond("a bearer token is required for pushing a wordbubble", http.StatusUnauthorized, w)
		return
	}

	token := splitToken[1] // grab the token from the Bearer string
	userId, err := app.auth.GetUserIdFromTokenString(token)
	if err != nil {
		app.respond(err.Error(), http.StatusUnauthorized, w)
		return
	}

	var wb WordBubble // finally we are authenticated! Let's insert a wordbubble
	if err = json.NewDecoder(r.Body).Decode(&wb); err != nil {
		logger.Error("could not decode wordbubble from body: %s", err)
		app.respond("could not parse a wordbubble from request body", http.StatusBadRequest, w)
		return
	}

	if err = app.wbs.ValidWordBubble(&wb); err != nil {
		app.respond(err.Error(), http.StatusBadRequest, w)
		return
	}

	if err = app.wbs.UserHasAvailability(userId); err != nil {
		app.respond(err.Error(), http.StatusConflict, w)
		return
	}

	if err = app.wbs.AddNewWordBubble(userId, &wb); err != nil {
		app.respond(err.Error(), http.StatusInternalServerError, w)
		return
	}

	if wb.Text == "teapot" {
		logger.Info("found ourselves a teapot")
		app.respond("here is some tea for you", http.StatusTeapot, w)
		return
	}

	logger.Info("successfully created a wordbubble for %d", userId)
	app.respond("thank you!", http.StatusCreated, w)
}

func (app *App) Pop(w http.ResponseWriter, r *http.Request) {
	logger := app.logger

	if r.Method != http.MethodDelete {
		logger.Error("invalid http method: %s", r.Method)
		app.respond("invalid http method", http.StatusMethodNotAllowed, w)
		return
	}

	var reqBody struct {
		UserStr string `json:"user"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		logger.Error("could not decode identity from body: %s", err)
		app.respond("could not parse a user from request body", http.StatusBadRequest, w)
		return
	}

	user := app.users.GetUserFromUserString(reqBody.UserStr)
	if user == nil {
		app.respond(fmt.Sprintf("could not resolve user `%s`", reqBody.UserStr), http.StatusBadRequest, w)
		return
	}

	wordbubble := app.wbs.RemoveAndReturnLatestWordBubbleForUserId(user.Id)
	if wordbubble == nil {
		app.respond(fmt.Sprintf("an internal error occurred while fetching wordbubble for %s", reqBody.UserStr), http.StatusInternalServerError, w)
		return
	}

	if wordbubble.Text == "" {
		app.respond(fmt.Sprintf("no wordbubble found for %s", reqBody.UserStr), http.StatusNoContent, w)
		return
	}

	logger.Info("successfully popped a wordbubble for %d", user.Id)
	app.respond(wordbubble.Text, http.StatusOK, w)
}
