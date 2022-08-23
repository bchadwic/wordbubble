package main

import (
	"encoding/json"
	"io"
	"net/http"
)

type App struct {
	logger Logger
	users  Users
}

func (app *App) SignUp(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("app.SignUp: handling request")

	if r.Method != http.MethodPost {
		app.logger.Warn("app.SignUp: invalid http method %s", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`invalid http method`))
		return // return since signup is only for posting a new user
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		app.logger.Error("app.SignUp: could not decode user from body %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`internal error occurred, could not retrieve user from request body`))
		return // could not parse a user from the request body
	}

	if !app.users.ValidPassword(user.Password) {
		app.logger.Error("app.SignUp: password complexity did not meet standard")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`password complexity is not sufficent, 1 upper case, 1 number, and >6 characters`))
		return // password complexity didn't match 1 upper case, 1 number and >6 characters
	}

	if err := app.users.AddUser(app.logger, &user); err != nil {
		app.logger.Error("app.SignUp: an error occurred inserting new user %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`internal error occurred, could not insert new user`))
		return // could not add a new user in the service or repository layer
	}

	app.logger.Info("app.SignUp: user %s created, returning token", user.Username)
	token, err := app.users.GenerateToken(app.logger, &user)
	if err != nil {
		app.logger.Error("app.SignUp: an error occurred generating token %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`internal error occurred, user has been inserted but could not generate token`))
		return // after adding the user, could not generate a token
	}

	app.logger.Info("app.SignUp: generated token was successful, sending back token response")
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
		w.Write([]byte(`internal error occurred, could not retrieve user from request body`))
		return // could not parse a user from the request body
	}

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
