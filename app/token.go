package app

import (
	"encoding/json"
	"net/http"

	"github.com/bchadwic/wordbubble/internal/auth"
	"github.com/bchadwic/wordbubble/resp"
)

func (wb *app) Token(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		wb.errorResponse(resp.ErrInvalidMethod, w)
		return
	}

	var reqBody struct {
		TokenString string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		wb.errorResponse(resp.ErrInvalidMethod, w)
		return
	}

	token, err := auth.RefreshTokenFromTokenString(reqBody.TokenString)
	if err != nil {
		wb.errorResponse(err, w)
		return
	}
	if err = wb.auth.ValidateRefreshToken(token); err != nil {
		wb.errorResponse(err, w)
		return
	}

	var latestRefreshToken string
	if token.IsNearEndOfLife() {
		latestRefreshToken, _ = wb.auth.GenerateRefreshToken(token.UserId())
	}

	resp := struct {
		RefreshToken string `json:"refresh_token,omitempty"`
		AccessToken  string `json:"access_token"`
	}{latestRefreshToken, wb.auth.GenerateAccessToken(token.UserId())}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
