package app

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/bchadwic/wordbubble/model/req"
	"github.com/bchadwic/wordbubble/model/resp"
	"github.com/bchadwic/wordbubble/util"
)

// Push queues a wordbubble for a user
// @Summary     Push a wordbubble
// @Description Push adds a new wordbubble to a user's queue
// @Tags        wordbubble
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       Wordbubble body     req.WordbubbleRequest true "Wordbubble containing the text to be stored"
// @Success     200        {object} resp.PushResponse
// @Failure     405        {object} resp.StatusMethodNotAllowed    "resp.ErrInvalidHttpMethod"
// @Failure     400        {object} resp.StatusBadRequest          "resp.ErrParseWordbubble, InvalidWordbubble"
// @Failure     409        {object} resp.StatusConflict            "resp.ErrMaxAmountOfWordbubblesReached"
// @Failure     401        {object} resp.StatusUnauthorized        "resp.ErrUnauthorized, resp.ErrInvalidTokenSignature"
// @Failure     500        {object} resp.StatusInternalServerError "resp.UnknownError"
// @Router      /push [post]
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

	var wordbubble req.WordbubbleRequest // finally we are authenticated! Let's insert a wordbubble
	if err = json.NewDecoder(r.Body).Decode(&wordbubble); err != nil {
		wb.errorResponse(resp.ErrParseWordbubble, w)
		return
	}

	err = wb.wordbubbles.AddNewWordbubble(userId, &wordbubble)
	if err != nil {
		wb.errorResponse(err, w)
		return
	}
	resp := &resp.PushResponse{
		Text: "thank you!",
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
