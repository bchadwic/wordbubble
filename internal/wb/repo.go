package wb

import (
	"database/sql"
	"errors"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/resp"
	"github.com/bchadwic/wordbubble/util"
	_ "github.com/mattn/go-sqlite3"
)

type wordBubbleRepo struct {
	db  *sql.DB
	log util.Logger
}

func NewWordBubbleRepo(logger util.Logger, db *sql.DB) *wordBubbleRepo {
	db.Exec(`
		PRAGMA foreign_keys = ON;
		CREATE TABLE IF NOT EXISTS wordbubbles (
			wordbubble_id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,  
			created_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			text TEXT NOT NULL,
			FOREIGN KEY(user_id) REFERENCES users(user_id)
		);
	`)
	return &wordBubbleRepo{log: logger, db: db}
}

type OutOfWordBubbleSlots error

func (repo *wordBubbleRepo) AddNewWordBubble(userId int64, wb *model.WordBubble) error {
	rs, err := repo.db.Exec(AddNewWordBubble, userId, wb.Text, userId, maxAmountOfWordBubbles)
	if err != nil {
		repo.log.Error("execute error for adding a wordbubble %+v for user: %d, error: %s", wb, userId, err)
		return err
	}
	amt, err := rs.RowsAffected()
	if err != nil {
		repo.log.Error("rows affected explosion %+v for user: %d, error: %s", wb, userId, err)
		return errors.New("couldn't determine how many wordbubbles you currently have")
	}
	if amt == 0 {
		return resp.ErrMaxAmountOfWordBubblesReached
	}
	return nil
}

func (repo *wordBubbleRepo) RemoveAndReturnLatestWordBubbleForUserId(userId int64) *model.WordBubble {
	row := repo.db.QueryRow(RemoveAndReturnLatestWordBubbleForUserId, userId)
	var wordbubble model.WordBubble
	if err := row.Scan(&wordbubble.Text); err != nil {
		repo.log.Error("could not map db wordbubble text for user: %d, error: %s", userId, err)
		return nil
	}
	return &wordbubble
}
