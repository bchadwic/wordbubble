package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type App struct {
	logger Logger
	auth   Auth
	users  Users
}

func (app *App) Register(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("app.Register: handling request")

	if r.Method != http.MethodPost {
		app.logger.Error("app.Register: invalid http method: %s", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`invalid http method`))
		return // return since Register is only for posting a new user
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		app.logger.Error("app.Register: could not decode user from body: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`could not retrieve user from request body`))
		return // could not parse a user from the request body
	}

	if err := app.users.ValidPassword(app.logger, user.Password); err != nil {
		app.logger.Error("app.Register: password complexity did not meet standard: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return // password complexity didn't match 1 upper case, 1 number and >6 characters
	}

	user.Username = strings.Trim(user.Username, " ")
	if err := app.users.ValidUser(app.logger, &user); err != nil {
		app.logger.Error("app.Register: could not add user: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return // either the username was taken, the email / username is invalid, or an account with this email exists
	}

	if err := app.users.AddUser(app.logger, &user); err != nil {
		app.logger.Error("app.Register: an error occurred inserting new user: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`internal error occurred, could not insert new user`))
		return // could not add a new user in the service or repository layer
	}

	app.logger.Info("app.Register: user %s created, returning token", user.Username)
	token, err := app.auth.GenerateToken(app.logger, &user)
	if err != nil {
		app.logger.Error("app.Register: an error occurred generating token: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`internal error occurred, user has been inserted but could not generate token`))
		return // after adding the user, could not generate a token
	}

	app.logger.Info("app.Register: generated token was successful, sending back token response")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(token)) // huzzah, another user added
}

func (app *App) Token(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("app.Token: handling request")

	if r.Method != http.MethodPost {
		app.logger.Error("app.Token: invalid http method: %s", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`invalid http method`))
		return // return since Token requires a body with credentials
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		app.logger.Error("app.Token: could not decode user from body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`could not retrieve user from request body`))
		return // could not parse a user from the request body
	}

	user.Username = strings.Trim(user.Username, " ")
	if ok := app.users.AuthenticateUser(app.logger, &user); !ok {
		app.logger.Error("app.Token: could not authenticate user")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`could not authenticate, please try again`))
		return // based on the credentials provided, could not authenticate
	}

	app.logger.Info("app.Token: user %s successfully logged in, returning token", user.Username)
	token, err := app.auth.GenerateToken(app.logger, &user)
	if err != nil {
		app.logger.Error("app.Token: an error occurred generating token: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`internal error occurred, user was authenticated but failed to make a token`))
		return // after adding the user, could not generate a token
	}

	app.logger.Info("app.Token: generated token was successful, sending back token response")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(token)) // twas a success
}

func (app *App) Push(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("app.Push: handling request")

	if r.Method != http.MethodPost {
		app.logger.Error("app.Push: invalid http method: %s", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`invalid http method`))
		return // return since Token requires a body with credentials
	}

	authValue := r.Header.Get("authorization")
	if authValue == "" {
		app.logger.Error("app.Push: authorization was not passed")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`authorization header is required for pushing a wordbubble`))
		return // return since user is not authenticated
	}

	splitToken := strings.Split(authValue, "Bearer ")
	if len(splitToken) < 2 {
		app.logger.Error("app.Push: authorization value didn't specifiy token type as bearer")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`a bearer token is required for pushing a wordbubble`))
		return // return since user is still not authenticated
	}

	token := splitToken[1] // grab the token from the Bearer string
	if err := app.auth.ValidateToken(app.logger, token); err != nil {
		app.logger.Error("app.Push: token is invalid: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return // return since user is, again, still not authenticated
	}

	// finally we are authenticated!
}
