package wb

import (
	cfg "github.com/bchadwic/wordbubble/internal/config"
	"github.com/bchadwic/wordbubble/model/req"
	"github.com/bchadwic/wordbubble/util"
)

type wordBubbleService struct {
	repo WordbubbleRepo
	log  util.Logger
}

func NewWordbubblesService(cfg cfg.Config, repo WordbubbleRepo) *wordBubbleService {
	return &wordBubbleService{
		log:  cfg.NewLogger("wordbubbles"),
		repo: repo,
	}
}

func (svc *wordBubbleService) AddNewWordbubble(userId int64, wb *req.Wordbubble) error {
	if err := util.ValidWordbubble(wb); err != nil {
		return err
	}
	return svc.repo.addNewWordbubble(userId, wb)
}

func (svc *wordBubbleService) RemoveAndReturnLatestWordbubbleForUserId(userId int64) *req.Wordbubble {
	return svc.repo.removeAndReturnLatestWordbubbleForUserId(userId)
}
