package wb

import (
	cfg "github.com/bchadwic/wordbubble/internal/config"
	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/util"
)

type wordBubbleService struct {
	repo WordBubbleRepo
	log  util.Logger
}

func NewWordBubblesService(cfg cfg.Config, repo WordBubbleRepo) *wordBubbleService {
	return &wordBubbleService{
		log:  cfg.NewLogger("wordbubbles"),
		repo: repo,
	}
}

func (svc *wordBubbleService) AddNewWordBubble(userId int64, wb *model.WordBubble) error {
	if err := util.ValidWordBubble(wb); err != nil {
		return err
	}
	return svc.repo.addNewWordBubble(userId, wb)
}

func (svc *wordBubbleService) RemoveAndReturnLatestWordBubbleForUserId(userId int64) *model.WordBubble {
	return svc.repo.removeAndReturnLatestWordBubbleForUserId(userId)
}
