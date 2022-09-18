package wb

import (
	"database/sql"
	"fmt"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/util"
	_ "github.com/mattn/go-sqlite3"
)

type WordBubbleRepo interface {
	AddNewWordBubble(userId int64, wb *model.WordBubble) error
	NumberOfWordBubblesForUser(userId int64) (int64, error)
	RemoveAndReturnLatestWordBubbleForUserId(userId int64) *model.WordBubble
}

type wordBubbleRepo struct {
	db  *sql.DB
	log util.Logger
}

func NewWordBubbleRepo(logger util.Logger) *wordBubbleRepo {
	panicker := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	db, err := sql.Open("sqlite3", "./wordbubble.db")
	panicker(err)
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS wordbubbles (
		wordbubble_id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,  
		created_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		text TEXT NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(user_id)
	);`)
	panicker(err)
	return &wordBubbleRepo{db, logger}
}

func (repo *wordBubbleRepo) AddNewWordBubble(userId int64, wb *model.WordBubble) error {
	stmt, err := repo.db.Prepare(`INSERT INTO wordbubbles (user_id, text) VALUES (?, ?)`)
	if err != nil {
		repo.log.Error("prepared statement error for adding a wordbubble %+v for user: %d, error: %s", wb, userId, err)
		return err
	}
	_, err = stmt.Exec(userId, wb.Text)
	if err != nil {
		repo.log.Error("execute error for adding a wordbubble %+v for user: %d, error: %s", wb, userId, err)
		return err
	}
	return nil
}

func (repo *wordBubbleRepo) NumberOfWordBubblesForUser(userId int64) (int64, error) {
	repo.log.Info("checking amount of wordbubbles for user %d", userId)
	stmt, err := repo.db.Prepare(`SELECT COUNT(*) from wordbubbles WHERE user_id = ?`)
	if err != nil {
		repo.log.Error("prepared statement error for getting number of wordbubbles for user: %d, error: %s", userId, err)
		return 0, err
	}
	row, err := stmt.Query(userId)
	if err != nil {
		repo.log.Error("query error for getting number of wordbubbles for user: %d, error: %s", userId, err)
		return 0, err
	}
	defer row.Close()
	if !row.Next() {
		repo.log.Error("could not find how many wordbubbles for user: %d", userId)
		return 0, fmt.Errorf("an error occurred gathering the current amount of wordbubbles")
	}
	var amt int64
	if err := row.Scan(&amt); err != nil {
		repo.log.Error("could not map db wordbubble amount for user: %s, error: %s", userId, err)
		return 0, fmt.Errorf("an error occurred gathering the current amount of wordbubbles for user")
	}
	return amt, nil
}

func (repo *wordBubbleRepo) RemoveAndReturnLatestWordBubbleForUserId(userId int64) *model.WordBubble {
	stmt, err := repo.db.Prepare(`
	DELETE FROM wordbubbles WHERE wordbubble_id = ( 
		SELECT wordbubble_id FROM wordbubbles WHERE user_id = ? ORDER BY created_timestamp ASC LIMIT 1
	) RETURNING text;`)
	if err != nil {
		repo.log.Error("prepared statement error for removing and returning wordbubble for user: %d, error: %s", userId, err)
		return nil
	}
	row, err := stmt.Query(userId)
	if err != nil {
		repo.log.Error("querying error for removing and returning wordbubble for user: %d, error: %s", userId, err)
		return nil
	}
	defer row.Close()
	if !row.Next() { // not logging because this might not be unexpected behaviour
		return nil
	}
	var wordbubble model.WordBubble
	if err := row.Scan(&wordbubble.Text); err != nil {
		repo.log.Error("could not map db wordbubble text for user: %s, error: %s", userId, err)
		return nil
	}
	return &wordbubble
}
