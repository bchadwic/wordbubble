package main

import "fmt"

const (
	minWordBubbleLength    = 1
	maxWordBubbleLength    = 255
	maxAmountOfWordBubbles = 10
)

type WordBubbles interface {
	AddNewWordBubble(logger Logger, userId int64, wb *WordBubble) error
	ValidWordBubble(wb *WordBubble) error
	UserHasAvailability(logger Logger, userId int64) error
}

type wordbubbles struct {
	ds DataSource
}

type WordBubble struct {
	Text string `json:"text"`
}

func NewWordBubblesService(ds DataSource) *wordbubbles {
	return &wordbubbles{ds}
}

func (wbs *wordbubbles) AddNewWordBubble(logger Logger, userId int64, wb *WordBubble) error {
	return wbs.ds.AddNewWordBubble(logger, userId, wb)
}

func (wbs *wordbubbles) UserHasAvailability(logger Logger, userId int64) error {
	logger.Info("wb.UserHasAvailability: processing")
	amt, err := wbs.ds.NumberOfWordBubblesForUser(logger, userId)
	if err != nil {
		return err
	}
	logger.Debug("wb.UserHasAvailability: successfully found %d wordbubbles for user %d", amt, userId)
	if amt >= maxAmountOfWordBubbles {
		return fmt.Errorf("you have exceeded the maximum amount of wordbubbles")
	}
	logger.Debug("wb.UserHasAvailability: successfully determined %d has room to add more wordbubbles", userId)
	return nil
}

func (wbs *wordbubbles) ValidWordBubble(wb *WordBubble) error {
	len := len(wb.Text)
	if len < minWordBubbleLength || len > maxWordBubbleLength {
		return fmt.Errorf("wordbubble sent is invalid, must be inbetween 1-255 characters, received %d", len)
	}
	return nil
}
