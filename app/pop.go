package app

import (
	"encoding/json"
	"net/http"

	"github.com/bchadwic/wordbubble/resp"
)

func (wb *app) Pop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		wb.errorResponse(resp.ErrInvalidHttpMethod, w)
		return
	}

	var reqBody struct {
		UserStr string `json:"user"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		wb.errorResponse(resp.ErrParseUser, w)
		return
	}

	user, err := wb.users.RetrieveUnauthenticatedUser(reqBody.UserStr)
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
