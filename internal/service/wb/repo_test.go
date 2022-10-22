package wb

import (
	"fmt"
	"testing"

	cfg "github.com/bchadwic/wordbubble/internal/config"
	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/resp"
	"github.com/stretchr/testify/assert"
)

func Test_HappyPath(t *testing.T) {
	repo := NewWordbubbleRepo(cfg.TestConfig())

	// A user gets created, now we can add a wordbubble
	_, err := repo.db.Exec(`INSERT INTO users (username, email, password) VALUES ('bchadwick', 'benchadwick87@gmail.com', 'test-password')`)
	if err != nil {
		panic(err)
	}
	// A user creates the max amount of wordbubbles
	for i := 0; i < maxAmountOfWordbubbles; i++ {
		err = repo.addNewWordbubble(1, &model.Wordbubble{
			Text: fmt.Sprintf("This is wordbubble #%d", i+1),
		})
		assert.Nil(t, err)
	}
	// A user tries to add one above the max amount, causing an error to be returned
	err = repo.addNewWordbubble(1, &model.Wordbubble{})
	assert.NotNil(t, err)
	assert.Error(t, resp.ErrMaxAmountOfWordbubblesReached, err)
	// A user wants space back so they start removing wordbubbles
	for i := 0; i < maxAmountOfWordbubbles; i++ {
		wordbubble := repo.removeAndReturnLatestWordbubbleForUserId(1)
		assert.Equal(t, fmt.Sprintf("This is wordbubble #%d", i+1), wordbubble.Text)
	}
	// A user tries to remove a non-existent wordbubble
	wordbubble := repo.removeAndReturnLatestWordbubbleForUserId(1)
	assert.Nil(t, wordbubble)
}
