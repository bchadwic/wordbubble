package app

import (
	"encoding/json"
	"net/http"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/resp"
	"github.com/bchadwic/wordbubble/util"
)

func (wb *app) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		wb.errorResponse(resp.ErrInvalidHttpMethod, w)
		return
	}

	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		wb.errorResponse(resp.ErrParseUser, w)
		return
	}

	if err := util.ValidUser(&user); err != nil {
		wb.errorResponse(err, w)
		return
	}

	if err := wb.users.AddUser(&user); err != nil {
		wb.errorResponse(err, w)
		return
	}

	refreshToken, err := wb.auth.GenerateRefreshToken(user.Id)
	if err != nil {
		wb.errorResponse(err, w)
		return
	}

	resp := struct {
		RefreshToken string `json:"refresh_token"`
		AccessToken  string `json:"access_token"`
	}{refreshToken, wb.auth.GenerateAccessToken(user.Id)}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
