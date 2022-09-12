package main

import "fmt"

const (
	minWordBubbleLength    = 1
	maxWordBubbleLength    = 255
	maxAmountOfWordBubbles = 10
)

type WordBubbles interface {
	AddNewWordBubble(userId int64, wb *WordBubble) error
	ValidWordBubble(wb *WordBubble) error
	UserHasAvailability(userId int64) error
	RemoveAndReturnLatestWordBubbleForUser(userId int64) (*WordBubble, error)
}

type wordbubbles struct {
	source DataSource
	log    Logger
}

type WordBubble struct {
	Text string `json:"text"`
}

func NewWordBubblesService(source DataSource, logger Logger) *wordbubbles {
	return &wordbubbles{source, logger}
}

func (wbs *wordbubbles) AddNewWordBubble(userId int64, wb *WordBubble) error {
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

func (wbs *wordbubbles) ValidWordBubble(wb *WordBubble) error {
	len := len(wb.Text)
	if len < minWordBubbleLength || len > maxWordBubbleLength {
		return fmt.Errorf("wordbubble sent is invalid, must be inbetween 1-255 characters, received %d", len)
	}
	return nil
}

func (wbs *wordbubbles) RemoveAndReturnLatestWordBubbleForUser(userId int64) (*WordBubble, error) {
	return wbs.source.RemoveAndReturnLatestWordBubbleForUser(userId)
}
