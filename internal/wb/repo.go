package wb

import (
	"database/sql"
	"fmt"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/util"
	_ "github.com/mattn/go-sqlite3"
)

type wordBubbleRepo struct {
	db  *sql.DB
	log util.Logger
}

func NewWordBubbleRepo(logger util.Logger, db *sql.DB) *wordBubbleRepo {
	db.Exec(`
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

func (repo *wordBubbleRepo) AddNewWordBubble(userId int64, wb *model.WordBubble) error {
	_, err := repo.db.Exec(AddNewWordBubble, userId, wb.Text)
	if err != nil {
		repo.log.Error("execute error for adding a wordbubble %+v for user: %d, error: %s", wb, userId, err)
		return err
	}
	return nil
}

func (repo *wordBubbleRepo) NumberOfWordBubblesForUser(userId int64) (int64, error) {
	row := repo.db.QueryRow(NumberOfWordBubblesForUser, userId)
	var amt int64
	if err := row.Scan(&amt); err != nil {
		repo.log.Error("could not map db wordbubble amount for user: %d, error: %s", userId, err)
		return 0, fmt.Errorf("an error occurred gathering the current amount of wordbubbles for user")
	}
	return amt, nil
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
