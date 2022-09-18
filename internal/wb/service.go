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

type WordBubbleService interface {
	AddNewWordBubble(userId int64, wb *model.WordBubble) error
	ValidWordBubble(wb *model.WordBubble) error
	UserHasAvailability(userId int64) error
	RemoveAndReturnLatestWordBubbleForUserId(userId int64) *model.WordBubble
}

type wordBubbleService struct {
	repo WordBubbleRepo
	log  util.Logger
}

func NewWordBubblesService(log util.Logger, repo WordBubbleRepo) *wordBubbleService {
	return &wordBubbleService{
		repo: repo,
		log:  log,
	}
}

func (svc *wordBubbleService) AddNewWordBubble(userId int64, wb *model.WordBubble) error {
	return svc.repo.AddNewWordBubble(userId, wb)
}

func (svc *wordBubbleService) UserHasAvailability(userId int64) error {
	svc.log.Info("processing")
	amt, err := svc.repo.NumberOfWordBubblesForUser(userId)
	if err != nil {
		return err
	}
	svc.log.Debug("successfully found %d wordbubbles for user %d", amt, userId)
	if amt >= maxAmountOfWordBubbles {
		return fmt.Errorf("you have exceeded the maximum amount of wordbubbles")
	}
	svc.log.Debug("successfully determined %d has room to add more wordbubbles", userId)
	return nil
}

func (svc *wordBubbleService) ValidWordBubble(wb *model.WordBubble) error {
	len := len(wb.Text)
	if len < minWordBubbleLength || len > maxWordBubbleLength {
		return fmt.Errorf("wordbubble sent is invalid, must be inbetween 1-255 characters, received %d", len)
	}
	return nil
}

func (wbs *wordBubbleService) RemoveAndReturnLatestWordBubbleForUserId(userId int64) *model.WordBubble {
	return wbs.repo.RemoveAndReturnLatestWordBubbleForUserId(userId)
}
