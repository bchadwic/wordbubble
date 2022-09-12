package wb

import (
	"fmt"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/util"
)

const (
	minWordBubbleLength    = 1
	maxWordBubbleLength    = 255
	maxAmountOfWordBubbles = 10
)

type WordBubbles interface {
	AddNewWordBubble(userId int64, wb *model.WordBubble) error
	ValidWordBubble(wb *model.WordBubble) error
	UserHasAvailability(userId int64) error
	RemoveAndReturnLatestWordBubbleForUserId(userId int64) *model.WordBubble
}

type wordbubbles struct {
	source DataSource
	log    util.Logger
}

func NewWordBubblesService(source DataSource, logger util.Logger) *wordbubbles {
	return &wordbubbles{source, logger}
}

func (wbs *wordbubbles) AddNewWordBubble(userId int64, wb *model.WordBubble) error {
	return wbs.source.AddNewWordBubble(userId, wb)
}

func (wbs *wordbubbles) UserHasAvailability(userId int64) error {
	wbs.log.Info("processing")
	amt, err := wbs.source.NumberOfWordBubblesForUser(userId)
	if err != nil {
		return err
	}
	wbs.log.Debug("successfully found %d wordbubbles for user %d", amt, userId)
	if amt >= maxAmountOfWordBubbles {
		return fmt.Errorf("you have exceeded the maximum amount of wordbubbles")
	}
	wbs.log.Debug("successfully determined %d has room to add more wordbubbles", userId)
	return nil
}

func (wbs *wordbubbles) ValidWordBubble(wb *model.WordBubble) error {
	len := len(wb.Text)
	if len < minWordBubbleLength || len > maxWordBubbleLength {
		return fmt.Errorf("wordbubble sent is invalid, must be inbetween 1-255 characters, received %d", len)
	}
	return nil
}

func (wbs *wordbubbles) RemoveAndReturnLatestWordBubbleForUserId(userId int64) *model.WordBubble {
	return wbs.source.RemoveAndReturnLatestWordBubbleForUserId(userId)
}
