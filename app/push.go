package app

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/resp"
	"github.com/bchadwic/wordbubble/util"
)

func (wb *app) Push(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		wb.errorResponse(resp.ErrInvalidHttpMethod, w)
		return
	}

	splitToken := strings.Split(r.Header.Get("authorization"), "Bearer ")
	if len(splitToken) < 2 {
		wb.errorResponse(resp.ErrUnauthorized, w)
		return
	}

	tokenStr := splitToken[1] // grab the token from the Bearer string
	userId, err := util.GetUserIdFromTokenString(tokenStr)
	if err != nil {
		wb.errorResponse(err, w)
		return
	}

	var wordbubble model.WordBubble // finally we are authenticated! Let's insert a wordbubble
	if err = json.NewDecoder(r.Body).Decode(&wordbubble); err != nil {
		wb.errorResponse(resp.ErrParseWordBubble, w)
		return
	}

	err = wb.wordbubbles.AddNewWordBubble(userId, &wordbubble)
	if err != nil {
		wb.errorResponse(err, w)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("thank you!"))
}
