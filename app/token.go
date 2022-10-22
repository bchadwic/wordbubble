package app

import (
	"encoding/json"
	"net/http"

	"github.com/bchadwic/wordbubble/internal/service/auth"
	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/resp"
)

// Token is used to retrieve a new access token from a refresh token
// @Summary     Token to api.wordbubble.io
// @Description Token to api.wordbubble.io for authorized use
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       Token body     model.RefreshToken true "Valid refresh token to gain a new access token"
// @Success     200   {object} resp.TokenResponse
// @Failure     400   {object} resp.StatusBadRequest       "resp.ErrParseRefreshToken"
// @Failure     401   {object} resp.StatusUnauthorized     "resp.ErrRefreshTokenIsExpired, resp.ErrCouldNotValidateRefreshToken"
// @Failure     405   {object} resp.StatusMethodNotAllowed "resp.ErrInvalidHttpMethod"
// @Failure     500   {object} resp.StatusMethodNotAllowed "resp.ErrCouldNotStoreRefreshToken"
// @Router      /token [post]
func (wb *app) Token(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		wb.errorResponse(resp.ErrInvalidHttpMethod, w)
		return
	}

	var reqBody model.RefreshToken
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		wb.errorResponse(resp.ErrParseRefreshToken, w)
		return
	}

	token, err := auth.RefreshTokenFromTokenString(reqBody.Token)
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

	resp := &resp.TokenResponse{
		RefreshToken: latestRefreshToken,
		AccessToken:  wb.auth.GenerateAccessToken(token.UserId()),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
