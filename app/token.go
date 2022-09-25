package app

import (
	"encoding/json"
	"net/http"

	"github.com/bchadwic/wordbubble/internal/auth"
	"github.com/bchadwic/wordbubble/resp"
)

func (app *App) Token(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.errorResponse(resp.ErrInvalidMethod, w)
		return
	}

	var reqBody struct {
		TokenString string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		app.errorResponse(resp.ErrInvalidMethod, w)
		return
	}

	token, err := auth.RefreshTokenFromTokenString(reqBody.TokenString)
	if err != nil {
		app.errorResponse(err, w)
		return
	}
	if err = app.auth.ValidateRefreshToken(token); err != nil {
		app.errorResponse(err, w)
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
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
