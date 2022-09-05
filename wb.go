package main

import "fmt"

const (
	minWordBubbleLength    = 1
	maxWordBubbleLength    = 255
	maxAmountOfWordBubbles = 10
)

type WordBubbles interface {
	AddNewWordBubble(logger Logger, username string, wb *WordBubble) error
	ValidWordBubble(wb *WordBubble) error
	UserHasAvailability(logger Logger, username string) error
}

type wordbubbles struct {
	db DataSource
}

type WordBubble struct {
	Text string `json:"text"`
}

func NewWordBubblesService() *wordbubbles {
	return &wordbubbles{}
}

func (wbs *wordbubbles) AddNewWordBubble(logger Logger, username string, wb *WordBubble) error {
	return wbs.db.AddNewWordBubble(logger, username, wb)
}

func (wbs *wordbubbles) UserHasAvailability(logger Logger, username string) error {
	amt, err := wbs.db.NumberOfWordBubblesForUser(logger, username)
	if err != nil {
		return err
	}
	if amt >= maxAmountOfWordBubbles {
		return fmt.Errorf("you have exceeded the maximum amount of wordbubbles")
	}
	return nil
}

func (wbs *wordbubbles) ValidWordBubble(wb *WordBubble) error {
	len := len(wb.Text)
	if len < minWordBubbleLength || len > maxWordBubbleLength {
		return fmt.Errorf("wordbubble sent is invalid, must be inbetween 1-255 characters, received %d", len)
	}
	return nil
}
