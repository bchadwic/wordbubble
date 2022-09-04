package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type App struct {
	logger Logger
	users  Users
}

func (app *App) Signup(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("app.Signup: handling request")

	if r.Method != http.MethodPost {
		app.logger.Warn("app.Signup: invalid http method %s", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`invalid http method`))
		return // return since Signup is only for posting a new user
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		app.logger.Error("app.Signup: could not decode user from body %s", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`could not retrieve user from request body`))
		return // could not parse a user from the request body
	}

	if err := app.users.ValidPassword(app.logger, user.Password); err != nil {
		app.logger.Error("app.Signup: password complexity did not meet standard, %s", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return // password complexity didn't match 1 upper case, 1 number and >6 characters
	}

	user.Username = strings.Trim(user.Username, " ")
	if err := app.users.ValidUser(app.logger, &user); err != nil {
		app.logger.Error("app.Signup: could not add user, %s", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return // either the username was taken, the email / username is invalid, or an account with this email exists
	}

	if err := app.users.AddUser(app.logger, &user); err != nil {
		app.logger.Error("app.Signup: an error occurred inserting new user %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`internal error occurred, could not insert new user`))
		return // could not add a new user in the service or repository layer
	}

	app.logger.Info("app.Signup: user %s created, returning token", user.Username)
	token, err := app.users.GenerateToken(app.logger, &user)
	if err != nil {
		app.logger.Error("app.Signup: an error occurred generating token %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`internal error occurred, user has been inserted but could not generate token`))
		return // after adding the user, could not generate a token
	}

	app.logger.Info("app.Signup: generated token was successful, sending back token response")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(token)) // huzzah, another user added
}

func (app *App) Login(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("app.Login: handling request")

	if r.Method != http.MethodPost {
		app.logger.Warn("app.Login: invalid http method %s", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`invalid http method`))
		return // return since login requires a body with credentials
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		app.logger.Error("app.Login: could not decode user from body %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`could not retrieve user from request body`))
		return // could not parse a user from the request body
	}

	user.Username = strings.Trim(user.Username, " ")
	if ok := app.users.AuthenticateUser(app.logger, &user); !ok {
		app.logger.Error("app.Login: could not authenticate user")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`could not authenticate, please try again`))
		return // based on the credentials provided, could not authenticate
	}

	app.logger.Info("app.Login: user %s successfully logged in, returning token", user.Username)
	token, err := app.users.GenerateToken(app.logger, &user)
	if err != nil {
		app.logger.Error("app.Login: an error occurred generating token %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`internal error occurred, user was authenticated but failed to make a token`))
		return // after adding the user, could not generate a token
	}

	app.logger.Info("app.Login: generated token was successful, sending back token response")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(token)) // twas a success
}

func (app *App) Latest(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("calling login")
	if r.Method == http.MethodGet {
		app.logger.Info("method was a get!!")
	} else if r.Method == http.MethodPost {
		app.logger.Info("method was a post!!")
	}
	io.WriteString(w, "Hi!")
}
