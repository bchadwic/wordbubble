package user

import (
	"testing"

	cfg "github.com/bchadwic/wordbubble/internal/config"
	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/model/resp"
	"github.com/stretchr/testify/assert"
)

func Test_HappyPath(t *testing.T) {
	repo := NewUserRepo(cfg.TestConfig())

	expected := &model.User{
		Id:       1,
		Username: "ben",
		Password: "test-password",
		Email:    "benchadwick87@gmail.com",
	}

	// someone signs up as a user
	actualId, err := repo.addUser(expected)
	assert.NoError(t, err)
	assert.Equal(t, expected.Id, actualId)

	// user needs to get a token with a user id from username
	actual, err := repo.retrieveUserByUsername("ben")
	assert.Nil(t, err)
	assert.Equal(t, expected.Id, actual.Id)
	assert.Equal(t, expected.Username, actual.Username)
	assert.Equal(t, expected.Email, actual.Email)
	assert.Equal(t, expected.Password, actual.Password)

	// same as previous step but with an email
	actual, err = repo.retrieveUserByEmail("benchadwick87@gmail.com")
	assert.Nil(t, err)
	assert.Equal(t, expected.Id, actual.Id)
	assert.Equal(t, expected.Username, actual.Username)
	assert.Equal(t, expected.Email, actual.Email)
	assert.Equal(t, expected.Password, actual.Password)

	// misstyped my email logining in
	actual, err = repo.retrieveUserByEmail("benchadwic87@gmail.com")
	assert.NotNil(t, err)
	assert.ErrorIs(t, resp.ErrUnknownUser, err)
	assert.Nil(t, actual)
}

func Test_NotSoHappyPath(t *testing.T) {
	repo := NewUserRepo(cfg.TestConfig())

	// simulate a mapping error between database and application
	repo.addUser(&model.User{
		Username: "ben",
	})
	_, err := repo.db.Exec(`ALTER TABLE users RENAME COLUMN username TO user_name;`)
	if err != nil {
		panic(err)
	}
	user, err := repo.retrieveUserByUsername("ben")
	assert.Nil(t, user)
	assert.NotNil(t, err)
	assert.ErrorIs(t, resp.ErrSQLMappingError, err)

	repo.db.Close()
	// an error occurs while adding a user
	id, err := repo.addUser(&model.User{})
	assert.Equal(t, int64(0), id)
	assert.NotNil(t, err)
	assert.ErrorIs(t, resp.ErrCouldNotAddUser, err)
}
