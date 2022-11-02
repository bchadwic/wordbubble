package wb

import (
	"database/sql"

	cfg "github.com/bchadwic/wordbubble/internal/config"
	"github.com/bchadwic/wordbubble/model/req"
	"github.com/bchadwic/wordbubble/model/resp"
	"github.com/bchadwic/wordbubble/util"
)

type wordBubbleRepo struct {
	db  *sql.DB
	log util.Logger
}

func NewWordbubbleRepo(config cfg.Config) *wordBubbleRepo {
	return &wordBubbleRepo{
		log: config.NewLogger("wb_repo"),
		db:  config.DB(),
	}
}

func (repo *wordBubbleRepo) addNewWordbubble(userId int64, wb *req.WordbubbleRequest) error {
	rs, err := repo.db.Exec(AddNewWordbubble, userId, wb.Text, userId, maxAmountOfWordbubbles)
	if err != nil {
		repo.log.Error("execute error for adding a wordbubble %+v for user: %d, error: %s", wb, userId, err)
		return err
	}
	amt, _ := rs.RowsAffected()
	if amt <= 0 {
		return resp.ErrMaxAmountOfWordbubblesReached
	}
	return nil
}

func (repo *wordBubbleRepo) removeAndReturnLatestWordbubbleForUserId(userId int64) *resp.WordbubbleResponse {
	row := repo.db.QueryRow(RemoveAndReturnLatestWordbubbleForUserId, userId)
	var wordbubble resp.WordbubbleResponse
	if err := row.Scan(&wordbubble.Text); err != nil {
		repo.log.Error("could not map db wordbubble text for user: %d, error: %s", userId, err)
		return nil
	}
	return &wordbubble
}
