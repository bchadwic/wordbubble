package wb

import (
	"github.com/bchadwic/wordbubble/model"
)

const (
	maxAmountOfWordBubbles = 10

	AddNewWordBubble                         = `INSERT INTO wordbubbles (user_id, text) SELECT ?, ? WHERE (SELECT COUNT(*) from wordbubbles WHERE user_id = ?) < ?;`
	RemoveAndReturnLatestWordBubbleForUserId = `DELETE FROM wordbubbles WHERE wordbubble_id = (SELECT wordbubble_id FROM wordbubbles WHERE user_id = ? ORDER BY created_timestamp ASC LIMIT 1) RETURNING text;`
)

type WordBubbleService interface {
	AddNewWordBubble(userId int64, wb *model.WordBubble) error
	RemoveAndReturnLatestWordBubbleForUserId(userId int64) *model.WordBubble
}

type WordBubbleRepo interface {
	AddNewWordBubble(userId int64, wb *model.WordBubble) error
	RemoveAndReturnLatestWordBubbleForUserId(userId int64) *model.WordBubble
}
