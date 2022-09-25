package app

import (
	"encoding/json"
	"net/http"

	"github.com/bchadwic/wordbubble/resp"
)

func (app *App) Pop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		app.errorResponse(resp.ErrInvalidMethod, w)
		return
	}

	var reqBody struct {
		UserStr string `json:"user"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		app.errorResponse(resp.ErrParseUser, w)
		return
	}

	user := app.users.RetrieveUserByString(reqBody.UserStr)
	if user == nil {
		app.errorResponse(resp.ErrUnknownUser, w)
		return
	}

	wordbubble := app.wordbubbles.RemoveAndReturnLatestWordBubbleForUserId(user.Id)
	if wordbubble == nil {
		app.errorResponse(resp.ErrNoWordBubble, w)
		return
	}

	writeResponse(w, http.StatusOK, wordbubble)
}
