package wb

import (
	"github.com/bchadwic/wordbubble/model"
)

const (
	maxAmountOfWordBubbles                   = 10
	AddNewWordBubble                         = `INSERT INTO wordbubbles (user_id, text) SELECT ?, ? WHERE (SELECT COUNT(*) from wordbubbles WHERE user_id = ?) < ?;`
	RemoveAndReturnLatestWordBubbleForUserId = `DELETE FROM wordbubbles WHERE wordbubble_id = (SELECT wordbubble_id FROM wordbubbles WHERE user_id = ? ORDER BY created_timestamp ASC LIMIT 1) RETURNING text;`
)

type WordBubbleService interface {
	// add a wordbubble for the user specified
	// an error is returned if wordbubble is not valid
	AddNewWordBubble(userId int64, wb *model.WordBubble) error
	// remove and returns the latest wordbubble for the user specified
	// the returned wordbubble may be null if none were found in the data source
	RemoveAndReturnLatestWordBubbleForUserId(userId int64) *model.WordBubble
}

type WordBubbleRepo interface {
	// add a validated wordbubble for the user specified
	AddNewWordBubble(userId int64, wb *model.WordBubble) error
	// remove and returns the latest wordbubble for the user specified
	RemoveAndReturnLatestWordBubbleForUserId(userId int64) *model.WordBubble
}
