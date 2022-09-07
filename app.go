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
	wbs    WordBubbles
}

func (app *App) Register(w http.ResponseWriter, r *http.Request) {
	logger := app.logger
	logger.Info("app.Register: handling request")

	if r.Method != http.MethodPost {
		logger.Error("app.Register: invalid http method: %s", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`invalid http method`))
		return // return since Register is only for posting a new user
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		logger.Error("app.Register: could not decode user from body: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`could not parse a user from request body`))
		return // return since we couldn't parse the body
	}

	if err := app.users.ValidPassword(logger, user.Password); err != nil {
		logger.Error("app.Register: password complexity did not meet standard: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return // password complexity didn't match 1 upper case, 1 number and >6 characters
	}

	user.Username = strings.Trim(user.Username, " ")
	if err := app.users.ValidUser(logger, &user); err != nil {
		logger.Error("app.Register: could not add user: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return // either the username was taken, the email / username is invalid, or an account with this email exists
	}

	if err := app.users.AddUser(logger, &user); err != nil {
		logger.Error("app.Register: an error occurred inserting new user: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`internal error occurred, could not insert new user`))
		return // could not add a new user in the service or repository layer
	}

	logger.Info("app.Register: user %s created, returning token", user.Username)
	token, err := app.auth.GenerateToken(logger, &user)
	if err != nil {
		logger.Error("app.Register: an error occurred generating token: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`internal error occurred, user has been inserted but could not generate token`))
		return // after adding the user, could not generate a token
	}

	logger.Info("app.Register: generated token was successful, sending back token response")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(token)) // huzzah, another user added
}

func (app *App) Token(w http.ResponseWriter, r *http.Request) {
	logger := app.logger
	logger.Info("app.Token: handling request")

	if r.Method != http.MethodPost {
		logger.Error("app.Token: invalid http method: %s", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`invalid http method`))
		return // return since Token requires a body with credentials
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		logger.Error("app.Token: could not decode user from body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`could not retrieve user from request body`))
		return // could not parse a user from the request body
	}

	user.Username = strings.Trim(user.Username, " ")
	if ok := app.users.AuthenticateUser(logger, &user); !ok {
		logger.Error("app.Token: could not authenticate user")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`could not authenticate, please try again`))
		return // based on the credentials provided, could not authenticate
	}

	logger.Info("app.Token: user %s successfully logged in, returning token", user.Username)
	token, err := app.auth.GenerateToken(logger, &user)
	if err != nil {
		logger.Error("app.Token: an error occurred generating token: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`internal error occurred, user was authenticated but failed to make a token`))
		return // after adding the user, could not generate a token
	}

	logger.Info("app.Token: generated token was successful, sending back token response")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(token)) // twas a success
}

func (app *App) Push(w http.ResponseWriter, r *http.Request) {
	logger := app.logger
	logger.Info("app.Push: handling request")

	if r.Method != http.MethodPost {
		logger.Error("app.Push: invalid http method: %s", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`invalid http method`))
		return // return since Token requires a body with credentials
	}
	logger.Debug("app.Push: method is POST")

	authValue := r.Header.Get("authorization")
	if authValue == "" {
		logger.Error("app.Push: authorization was not passed")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`authorization header is required for pushing a wordbubble`))
		return // return since user is not authenticated
	}
	logger.Debug("app.Push: user passed authorization in header")

	splitToken := strings.Split(authValue, "Bearer ")
	if len(splitToken) < 2 {
		logger.Error("app.Push: authorization value didn't specifiy token type as bearer")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`a bearer token is required for pushing a wordbubble`))
		return // return since user is still not authenticated
	}
	logger.Debug("app.Push: user passed a token of type bearer")

	token := splitToken[1] // grab the token from the Bearer string
	userId, err := app.auth.ValidateTokenAndReceiveId(logger, token)
	if err != nil {
		logger.Error("app.Push: token is invalid: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return // return since user is, again, still not authenticated
	}
	logger.Debug("app.Push: successfully parsed and validated token")

	var wb WordBubble // finally we are authenticated! Let's insert a wordbubble
	if err = json.NewDecoder(r.Body).Decode(&wb); err != nil {
		logger.Error("app.Push: could not decode wordbubble from body: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`could not parse a wordbubble from request body`))
		return // return since we couldn't parse the body
	}
	logger.Debug("app.Push: successfully parsed wordbubble from request body")

	if err = app.wbs.ValidWordBubble(&wb); err != nil {
		logger.Error("app.Push: wordbubble was not valid: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return // return since the wordbubble wasn't valid
	}
	logger.Debug("app.Push: successfully validated wordbubble passed")

	if err = app.wbs.UserHasAvailability(logger, userId); err != nil {
		logger.Error("app.Push: user doesn't have availability for another wb: %s", err)
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(err.Error()))
		return // return user has exceeded the amount of available wordbubbles
	}
	logger.Debug("app.Push: %d has availablity", userId)

	if err = app.wbs.AddNewWordBubble(logger, userId, &wb); err != nil {
		logger.Error("app.Push: there was an error creating the wordbubble: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return // returning due to internal error
	}
	logger.Debug("app.Push: %d added a wordbubble", userId)

	if wb.Text == "teapot" {
		logger.Info("app.Push: found ourselves a teapot")
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(`here is some tea for you`))
		return // return for obvious reasons
	}

	logger.Info("app.Push: successfully created a wordbubble for %d", userId)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("thank you!"))
}

func (app *App) Pop(w http.ResponseWriter, r *http.Request) {
	logger := app.logger

	if r.Method != http.MethodDelete {
		logger.Error("app.Pop: invalid http method: %s", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`invalid http method`))
		return // return since pop is a delete operation
	}

	var identity struct {
		UserStr string `json:"user"`
	}
	if err := json.NewDecoder(r.Body).Decode(&identity); err != nil {
		logger.Error("app.Pop: could not decode identity from body: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`could not parse a user from request body`))
		return // return since we couldn't parse the body
	}

	userId, err := app.users.ResolveUserIdFromValue(logger, identity.UserStr)
	if err != nil {
		logger.Error("app.Pop: an error occurred resolving the identity of %s: %s", identity.UserStr, err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return // return since we couldn't find who we were asking for
	}

	wordbubble, err := app.wbs.RemoveAndReturnLatestWordBubbleForUser(logger, userId)
	if err != nil {
		logger.Error("app.Pop: an error occurred removing and return the latest wordbubble for %d: %s", userId, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return // return since we had an internal error trying to pop
	}

	if wordbubble == nil {
		logger.Warn("app.Pop: no wordbubble was found for user %d", userId)
		w.WriteHeader(http.StatusNoContent)
		return // return with no content since we had nothing to delete
	}

	logger.Info("app.Pop: successfully popped a wordbubble for %d", userId)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(wordbubble.Text))
}
