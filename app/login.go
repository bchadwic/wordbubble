package app

import (
	"encoding/json"
	"net/http"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/resp"
)

// Login is used to get the access and refresh token for a user's credentials
// @Summary     Login to api.wordbubble.io
// @Description Login to api.wordbubble.io using the user credentials
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       User body     model.LoginUser                true "Credentials used to authenticate a user"
// @Success     200  {object} resp.TokenResponse             "Valid access and refresh tokens for user"
// @Failure     405  {object} resp.StatusMethodNotAllowed    "resp.ErrInvalidHttpMethod"
// @Failure     400  {object} resp.StatusBadRequest          "resp.ErrParseUser, resp.ErrUnknownUser, resp.ErrCouldNotDetermineUserType"
// @Failure     401  {object} resp.StatusUnauthorized        "resp.ErrInvalidCredentials"
// @Failure     500  {object} resp.StatusInternalServerError "resp.ErrSQLMappingError, resp.ErrCouldNotStoreRefreshToken"
// @Router      /login [post]
func (wb *app) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		wb.errorResponse(resp.ErrInvalidHttpMethod, w)
		return
	}

	var user model.LoginUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		wb.errorResponse(resp.ErrParseUser, w)
		return
	}

	authenticatedUser, err := wb.users.RetrieveAuthenticatedUser(user.User, user.Password)
	if err != nil {
		wb.errorResponse(err, w)
		return
	}

	refreshToken, err := wb.auth.GenerateRefreshToken(authenticatedUser.Id)
	if err != nil {
		wb.errorResponse(err, w)
		return
	}

	resp := resp.TokenResponse{
		RefreshToken: refreshToken,
		AccessToken:  wb.auth.GenerateAccessToken(authenticatedUser.Id),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
