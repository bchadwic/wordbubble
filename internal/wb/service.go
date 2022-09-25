package wb

import (
	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/util"
)

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
	if err := util.ValidWordBubble(wb); err != nil {
		return err
	}
	return svc.repo.AddNewWordBubble(userId, wb)
}

func (svc *wordBubbleService) RemoveAndReturnLatestWordBubbleForUserId(userId int64) *model.WordBubble {
	return svc.repo.RemoveAndReturnLatestWordBubbleForUserId(userId)
}
