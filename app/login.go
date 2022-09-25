package app

import (
	"encoding/json"
	"net/http"

	"github.com/bchadwic/wordbubble/resp"
)

func (app *App) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.errorResponse(resp.ErrInvalidMethod, w)
		return
	}

	var reqBody struct {
		User     string `json:"user"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		app.errorResponse(resp.ErrParseUser, w)
		return
	}

	AuthenticateUser := app.users.RetrieveAuthenticatedUserByString(reqBody.User, reqBody.Password)
	if AuthenticateUser == nil {
		app.errorResponse(resp.ErrInvalidCredentials, w)
		return
	}

	refreshToken, err := app.auth.GenerateRefreshToken(AuthenticateUser.Id)
	if err != nil {
		app.errorResponse(err, w)
		return
	}

	resp := struct {
		RefreshToken string `json:"refresh_token"`
		AccessToken  string `json:"access_token"`
	}{refreshToken, app.auth.GenerateAccessToken(AuthenticateUser.Id)}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
