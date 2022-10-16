package wb

import (
	"github.com/bchadwic/wordbubble/model"
)

const (
	maxAmountOfWordBubbles                   = 10
	AddNewWordBubble                         = `INSERT INTO wordbubbles (user_id, text) SELECT ?, ? WHERE (SELECT COUNT(*) from wordbubbles WHERE user_id = ?) < ?;`
	RemoveAndReturnLatestWordBubbleForUserId = `DELETE FROM wordbubbles WHERE wordbubble_id = (SELECT wordbubble_id FROM wordbubbles WHERE user_id = ? ORDER BY created_timestamp ASC LIMIT 1) RETURNING text;`
)

// WordBubbleService is the interface that
// the application uses to interact with wordbubbles
type WordBubbleService interface {
	// AddNewWordBubble adds a wordbubble for the user specified
	// error can be (400) - invalid wordbubble - resp.BadRequest,
	// (409) resp.ErrMaxAmountOfWordBubblesReached, (500) resp.UnknownError or nil.
	AddNewWordBubble(userId int64, wb *model.WordBubble) error
	// RemoveAndReturnLatestWordBubbleForUserId remove and returns the latest wordbubble for the user specified.
	// *model.WordBubble may be nil if none were found in the data source.
	RemoveAndReturnLatestWordBubbleForUserId(userId int64) *model.WordBubble
}

// WordBubbleRepo is the interface that the
// service layer uses to interact with wordbubbles
type WordBubbleRepo interface {
	// addNewWordBubble adds a validated wordbubble for the user specified.
	// error can be (409) resp.ErrMaxAmountOfWordBubblesReached, (500) resp.UnknownError or nil.
	addNewWordBubble(userId int64, wb *model.WordBubble) error
	// removeAndReturnLatestWordBubbleForUserId remove and returns the latest wordbubble for the user specified, could be nil
	// *model.WordBubble may be nil if none were found in the data source.
	removeAndReturnLatestWordBubbleForUserId(userId int64) *model.WordBubble
}
