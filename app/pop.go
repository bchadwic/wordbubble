package app

import (
	"encoding/json"
	"net/http"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/resp"
)

// Pop removes and returns a wordbubble for a user
// @Summary     Pop a wordbubble
// @Description Pop removes and returns a wordbubble for a user
// @Tags        wordbubble
// @Accept      json
// @Produce     json
// @Param       UnauthenticatedUser body     model.PopUser true "Username or email that the wordbubble will come from"
// @Success     200  {object} model.WordBubble					"Latest WordBubble for user passed"
// @Success     201  {object} resp.StatusNoContent 				"resp.ErrNoWordBubble"
// @Failure     405  {object} resp.StatusMethodNotAllowed     	"resp.ErrInvalidHttpMethod"
// @Failure     400  {object} resp.StatusBadRequest             "resp.ErrParseUser, resp.ErrUnknownUser, resp.ErrCouldNotDetermineUserType"
// @Failure     401  {object} resp.StatusUnauthorized           "resp.ErrInvalidCredentials"
// @Failure     500  {object} resp.StatusInternalServerError 	"resp.ErrSQLMappingError, resp.ErrCouldNotStoreRefreshToken"
// @Router      /pop [delete]
func (wb *app) Pop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		wb.errorResponse(resp.ErrInvalidHttpMethod, w)
		return
	}

	var reqBody model.PopUser
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		wb.errorResponse(resp.ErrParseUser, w)
		return
	}

	user, err := wb.users.RetrieveUnauthenticatedUser(reqBody.User)
	if err != nil {
		wb.errorResponse(err, w)
		return
	}

	wordbubble := wb.wordbubbles.RemoveAndReturnLatestWordBubbleForUserId(user.Id)
	if wordbubble == nil {
		wb.errorResponse(resp.ErrNoWordBubble, w)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(wordbubble)
}
