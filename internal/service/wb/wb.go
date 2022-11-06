package wb

import (
	"github.com/bchadwic/wordbubble/model/req"
	"github.com/bchadwic/wordbubble/model/resp"
)

const (
	maxAmountOfWordbubbles                   = 10
	AddNewWordbubble                         = `INSERT INTO wordbubbles (user_id, text) SELECT $1, $2 WHERE (SELECT COUNT(*) from wordbubbles WHERE user_id = $3) < ?;`
	RemoveAndReturnLatestWordbubbleForUserId = `DELETE FROM wordbubbles WHERE wordbubble_id = (SELECT wordbubble_id FROM wordbubbles WHERE user_id = $1 ORDER BY created_timestamp ASC LIMIT 1) RETURNING text;`
)

// WordbubbleService is the interface that
// the application uses to interact with wordbubbles
type WordbubbleService interface {
	// AddNewWordbubble adds a wordbubble for the user specified
	// error can be (400) - invalid wordbubble - resp.BadRequest,
	// (409) resp.ErrMaxAmountOfWordbubblesReached, (500) resp.UnknownError or nil.
	AddNewWordbubble(userId int64, wb *req.WordbubbleRequest) error
	// RemoveAndReturnLatestWordbubbleForUserId remove and returns the latest wordbubble for the user specified.
	// *req.Wordbubble may be nil if none were found in the data source.
	RemoveAndReturnLatestWordbubbleForUserId(userId int64) *resp.WordbubbleResponse
}

// WordbubbleRepo is the interface that the
// service layer uses to interact with wordbubbles
type WordbubbleRepo interface {
	// addNewWordbubble adds a validated wordbubble for the user specified.
	// error can be (409) resp.ErrMaxAmountOfWordbubblesReached, (500) resp.UnknownError or nil.
	addNewWordbubble(userId int64, wb *req.WordbubbleRequest) error
	// removeAndReturnLatestWordbubbleForUserId remove and returns the latest wordbubble for the user specified, could be nil
	// *req.Wordbubble may be nil if none were found in the data source.
	removeAndReturnLatestWordbubbleForUserId(userId int64) *resp.WordbubbleResponse
}
