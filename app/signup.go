package app

import (
	"encoding/json"
	"net/http"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/model/req"
	"github.com/bchadwic/wordbubble/model/resp"
	"github.com/bchadwic/wordbubble/util"
)

// Signup is used to signup a new user
// @Summary     Signup to api.wordbubble.io
// @Description Signup to api.wordbubble.io using a unique email and username
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       User body     req.SignupUserRequest true "User information required to signup"
// @Success     200  {object} resp.TokenResponse
// @Failure     400  {object} resp.StatusBadRequest          "resp.ErrParseUser, resp.ErrEmailIsNotValid, resp.ErrEmailIsTooLong, resp.ErrUsernameIsTooLong, resp.ErrUsernameIsNotLongEnough, resp.ErrUsernameInvalidChars, resp.ErrUserWithUsernameAlreadyExists, resp.ErrUserWithEmailAlreadyExists, resp.ErrCouldNotDetermineUserExistence, InvalidPassword"
// @Failure     405  {object} resp.StatusMethodNotAllowed    "resp.ErrInvalidHttpMethod"
// @Failure     500  {object} resp.StatusInternalServerError "resp.ErrCouldNotBeHashPassword, resp.ErrCouldNotAddUser, resp.ErrCouldNotStoreRefreshToken"
// @Router      /signup [post]
func (wb *app) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		wb.errorResponse(resp.ErrInvalidHttpMethod, w)
		return
	}

	var reqUser req.SignupUserRequest
	if err := json.NewDecoder(r.Body).Decode(&reqUser); err != nil {
		wb.errorResponse(resp.ErrParseUser, w)
		return
	}

	user := &model.User{
		Username: reqUser.Username,
		Email:    reqUser.Email,
		Password: reqUser.Password,
	}
	if err := util.ValidUser(user); err != nil {
		wb.errorResponse(err, w)
		return
	}

	if err := wb.users.AddUser(user); err != nil {
		wb.errorResponse(err, w)
		return
	}

	refreshToken, err := wb.auth.GenerateRefreshToken(user.Id)
	if err != nil {
		wb.errorResponse(err, w)
		return
	}

	resp := &resp.TokenResponse{
		RefreshToken: refreshToken,
		AccessToken:  wb.auth.GenerateAccessToken(user.Id),
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
