package wb

import (
	"github.com/bchadwic/wordbubble/model"
)

const (
	maxAmountOfWordbubbles                   = 10
	AddNewWordbubble                         = `INSERT INTO wordbubbles (user_id, text) SELECT ?, ? WHERE (SELECT COUNT(*) from wordbubbles WHERE user_id = ?) < ?;`
	RemoveAndReturnLatestWordbubbleForUserId = `DELETE FROM wordbubbles WHERE wordbubble_id = (SELECT wordbubble_id FROM wordbubbles WHERE user_id = ? ORDER BY created_timestamp ASC LIMIT 1) RETURNING text;`
)

// WordbubbleService is the interface that
// the application uses to interact with wordbubbles
type WordbubbleService interface {
	// AddNewWordbubble adds a wordbubble for the user specified
	// error can be (400) - invalid wordbubble - resp.BadRequest,
	// (409) resp.ErrMaxAmountOfWordbubblesReached, (500) resp.UnknownError or nil.
	AddNewWordbubble(userId int64, wb *model.Wordbubble) error
	// RemoveAndReturnLatestWordbubbleForUserId remove and returns the latest wordbubble for the user specified.
	// *model.Wordbubble may be nil if none were found in the data source.
	RemoveAndReturnLatestWordbubbleForUserId(userId int64) *model.Wordbubble
}

// WordbubbleRepo is the interface that the
// service layer uses to interact with wordbubbles
type WordbubbleRepo interface {
	// addNewWordbubble adds a validated wordbubble for the user specified.
	// error can be (409) resp.ErrMaxAmountOfWordbubblesReached, (500) resp.UnknownError or nil.
	addNewWordbubble(userId int64, wb *model.Wordbubble) error
	// removeAndReturnLatestWordbubbleForUserId remove and returns the latest wordbubble for the user specified, could be nil
	// *model.Wordbubble may be nil if none were found in the data source.
	removeAndReturnLatestWordbubbleForUserId(userId int64) *model.Wordbubble
}
