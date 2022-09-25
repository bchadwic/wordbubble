package wb

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/resp"
	"github.com/bchadwic/wordbubble/util"
	"github.com/stretchr/testify/assert"
)

func NewTestDB() *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			user_id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL
		);
	`)
	return db
}

func Test_HappyPath(t *testing.T) {
	repo := NewWordBubbleRepo(util.TestLogger(), NewTestDB())

	// A non existent user adds a wordbubble
	err := repo.AddNewWordBubble(1, &model.WordBubble{})
	assert.NotNil(t, err)
	assert.Equal(t, "FOREIGN KEY constraint failed", err.Error())

	// A user gets created, now we can add a wordbubble
	_, err = repo.db.Exec(`INSERT INTO users (username, email, password) VALUES ('bchadwick', 'benchadwick87@gmail.com', 'test-password')`)
	if err != nil {
		panic(err)
	}
	// A user creates the max amount of wordbubbles
	for i := 0; i < maxAmountOfWordBubbles; i++ {
		err = repo.AddNewWordBubble(1, &model.WordBubble{
			Text: fmt.Sprintf("This is wordbubble #%d", i+1),
		})
		assert.Nil(t, err)
	}
	// A user tries to add one above the max amount, causing an error to be returned
	err = repo.AddNewWordBubble(1, &model.WordBubble{})
	assert.NotNil(t, err)
	assert.Error(t, resp.ErrMaxAmountOfWordBubblesReached, err)
	// A user wants space back so they start removing wordbubbles
	for i := 0; i < maxAmountOfWordBubbles; i++ {
		wordbubble := repo.RemoveAndReturnLatestWordBubbleForUserId(1)
		assert.Equal(t, fmt.Sprintf("This is wordbubble #%d", i+1), wordbubble.Text)
	}
	// A user tries to remove a non-existent wordbubble
	wordbubble := repo.RemoveAndReturnLatestWordBubbleForUserId(1)
	assert.Nil(t, wordbubble)
}
