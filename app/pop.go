package app

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/bchadwic/wordbubble/model/req"
	"github.com/bchadwic/wordbubble/model/resp"
)

// Pop removes and returns a wordbubble for a user
// @Summary     Pop a wordbubble
// @Description Pop removes and returns a wordbubble for a user
// @Tags        wordbubble
// @Accept      json
// @Produce     json
// @Param       UnauthenticatedUser body     req.PopUserRequest             true "Username or email that the wordbubble will come from"
// @Success     200                 {object} resp.WordbubbleResponse        "Latest Wordbubble for user passed"
// @Success     201                 {object} resp.StatusNoContent           "resp.ErrNoWordbubble"
// @Failure     405                 {object} resp.StatusMethodNotAllowed    "resp.ErrInvalidHttpMethod"
// @Failure     400                 {object} resp.StatusBadRequest          "resp.ErrParseUser, resp.ErrNoUser, resp.ErrUnknownUser, resp.ErrCouldNotDetermineUserType"
// @Failure     401                 {object} resp.StatusUnauthorized        "resp.ErrInvalidCredentials"
// @Failure     500                 {object} resp.StatusInternalServerError "resp.ErrSQLMappingError, resp.ErrCouldNotStoreRefreshToken"
// @Router      /pop [delete]
func (wb *app) Pop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		wb.errorResponse(resp.ErrInvalidHttpMethod, w)
		return
	}

	user, err := getPopUserFromBody(r.Body)
	if err != nil {
		wb.errorResponse(err, w)
		return
	}

	unauthenticatedUser, err := wb.users.RetrieveUnauthenticatedUser(user.User)
	if err != nil {
		wb.errorResponse(err, w)
		return
	}

	wordbubble := wb.wordbubbles.RemoveAndReturnLatestWordbubbleForUserId(unauthenticatedUser.Id)
	if wordbubble == nil {
		wb.errorResponse(resp.ErrNoWordbubble, w)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(wordbubble)
}

func getPopUserFromBody(body io.Reader) (*req.PopUserRequest, error) {
	var user req.PopUserRequest
	if err := json.NewDecoder(body).Decode(&user); err != nil {
		return nil, resp.ErrParseUser
	}
	if user.User == "" {
		return nil, resp.ErrNoUser
	}
	return &user, nil
}
