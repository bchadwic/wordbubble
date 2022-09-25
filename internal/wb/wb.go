package wb

import (
	"github.com/bchadwic/wordbubble/model"
)

const (
	minWordBubbleLength    = 1
	maxWordBubbleLength    = 255
	maxAmountOfWordBubbles = 10

	AddNewWordBubble                         = `INSERT INTO wordbubbles (user_id, text) VALUES (?, ?)`
	NumberOfWordBubblesForUser               = `SELECT COUNT(*) from wordbubbles WHERE user_id = ?`
	RemoveAndReturnLatestWordBubbleForUserId = `DELETE FROM wordbubbles WHERE wordbubble_id = (SELECT wordbubble_id FROM wordbubbles WHERE user_id = ? ORDER BY created_timestamp ASC LIMIT 1) RETURNING text;`
)

type WordBubbleService interface {
	AddNewWordBubble(userId int64, wb *model.WordBubble) error
	ValidWordBubble(wb *model.WordBubble) error
	UserHasAvailability(userId int64) error
	RemoveAndReturnLatestWordBubbleForUserId(userId int64) *model.WordBubble
}

type WordBubbleRepo interface {
	AddNewWordBubble(userId int64, wb *model.WordBubble) error
	NumberOfWordBubblesForUser(userId int64) (int64, error)
	RemoveAndReturnLatestWordBubbleForUserId(userId int64) *model.WordBubble
}
