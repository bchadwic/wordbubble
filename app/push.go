package app

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/resp"
	"github.com/bchadwic/wordbubble/util"
)

func (app *App) Push(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.errorResponse(resp.ErrInvalidMethod, w)
		return
	}

	splitToken := strings.Split(r.Header.Get("authorization"), "Bearer ")
	if len(splitToken) < 2 {
		app.errorResponse(resp.ErrUnauthorized, w)
		return
	}

	tokenStr := splitToken[1] // grab the token from the Bearer string
	userId, err := util.GetUserIdFromTokenString(tokenStr)
	if err != nil {
		app.errorResponse(err, w)
		return
	}

	var wb model.WordBubble // finally we are authenticated! Let's insert a wordbubble
	if err = json.NewDecoder(r.Body).Decode(&wb); err != nil {
		app.errorResponse(resp.ErrParseWordBubble, w)
		return
	}

	err = app.wordbubbles.AddNewWordBubble(userId, &wb)
	if err != nil {
		app.errorResponse(err, w)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("thank you!"))
}
