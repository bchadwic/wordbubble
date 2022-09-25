package app

import (
	"encoding/json"
	"net/http"

	"github.com/bchadwic/wordbubble/resp"
)

func (wb *app) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		wb.errorResponse(resp.ErrInvalidMethod, w)
		return
	}

	var reqBody struct {
		User     string `json:"user"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		wb.errorResponse(resp.ErrParseUser, w)
		return
	}

	AuthenticateUser := wb.users.RetrieveAuthenticatedUserByString(reqBody.User, reqBody.Password)
	if AuthenticateUser == nil {
		wb.errorResponse(resp.ErrInvalidCredentials, w)
		return
	}

	refreshToken, err := wb.auth.GenerateRefreshToken(AuthenticateUser.Id)
	if err != nil {
		wb.errorResponse(err, w)
		return
	}

	resp := struct {
		RefreshToken string `json:"refresh_token"`
		AccessToken  string `json:"access_token"`
	}{refreshToken, wb.auth.GenerateAccessToken(AuthenticateUser.Id)}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}