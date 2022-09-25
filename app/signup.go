package app

import (
	"encoding/json"
	"net/http"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/resp"
	"github.com/bchadwic/wordbubble/util"
)

func (app *App) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.errorResponse(resp.ErrInvalidMethod, w)
		return
	}

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		app.errorResponse(resp.ErrParseUser, w)
		return
	}

	if err := util.ValidUser(&user); err != nil {
		app.errorResponse(err, w)
		return
	}

	if err := app.users.AddUser(&user); err != nil {
		app.errorResponse(err, w)
		return
	}

	refreshToken, err := app.auth.GenerateRefreshToken(user.Id)
	if err != nil {
		app.errorResponse(err, w)
		return
	}

	resp := struct {
		RefreshToken string `json:"refresh_token"`
		AccessToken  string `json:"access_token"`
	}{refreshToken, app.auth.GenerateAccessToken(user.Id)}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
