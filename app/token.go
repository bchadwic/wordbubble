package app

import (
	"encoding/json"
	"net/http"

	"github.com/bchadwic/wordbubble/internal/service/auth"
	"github.com/bchadwic/wordbubble/resp"
)

// Token is used to retrieve a new access token from a refresh token
// @Summary     Token to api.wordbubble.io
// @Description Token to api.wordbubble.io for authorized use
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       User body     model.SignupUser true "User information required to signup"
// @Success     200  {object} 		string
// @Failure     400  {object} 		resp.StatusBadRequest			"resp.ErrParseUser, resp.ErrEmailIsNotValid, resp.ErrEmailIsTooLong, resp.ErrUsernameIsTooLong, resp.ErrUsernameIsNotLongEnough, resp.ErrUsernameInvalidChars, resp.ErrUserWithUsernameAlreadyExists, resp.ErrUserWithEmailAlreadyExists, resp.ErrCouldNotDetermineUserExistence, InvalidPassword"
// @Failure     405  {object} 		resp.StatusMethodNotAllowed		"resp.ErrInvalidHttpMethod"
// @Failure     500  {object} 		resp.StatusInternalServerError	"resp.ErrCouldNotBeHashPassword, resp.ErrCouldNotAddUser, resp.ErrCouldNotStoreRefreshToken"
// @Router      /token [post]
func (wb *app) Token(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		wb.errorResponse(resp.ErrInvalidHttpMethod, w)
		return
	}

	var reqBody struct {
		TokenString string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		wb.errorResponse(resp.ErrInvalidHttpMethod, w)
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
