package user

import (
	"testing"

	cfg "github.com/bchadwic/wordbubble/internal/config"
	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/model/resp"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func Test_AddUser(t *testing.T) {
	tests := map[string]struct {
		user        *model.User
		repo        *testUserRepo
		expectedErr error
	}{
		"valid": {
			user: &model.User{
				Id:       6,
				Password: "hello world",
			},
			repo: &testUserRepo{
				errRetrieveUser:  resp.ErrUnknownUser,
				errRetrieveEmail: resp.ErrUnknownUser,
			},
		},
		"db had a problem adding user": {
			user: &model.User{},
			repo: &testUserRepo{
				errRetrieveUser:  resp.ErrUnknownUser,
				errRetrieveEmail: resp.ErrUnknownUser,
				errAddUser:       resp.ErrCouldNotAddUser,
			},
			expectedErr: resp.ErrCouldNotAddUser,
		},
		"user already exists with username": {
			user: &model.User{},
			repo: &testUserRepo{
				userRetrieveUserByUsername: &model.User{},
			},
			expectedErr: resp.ErrUserWithUsernameAlreadyExists,
		},
		"user might exist with this username, not sure due to sql mapping": {
			user: &model.User{},
			repo: &testUserRepo{
				errRetrieveUser: resp.ErrSQLMappingError,
			},
			expectedErr: resp.ErrCouldNotDetermineUserExistence,
		},
		"user already exists with email": {
			user: &model.User{},
			repo: &testUserRepo{
				userRetrieveUserByEmail: &model.User{},
				errRetrieveUser:         resp.ErrUnknownUser,
			},
			expectedErr: resp.ErrUserWithEmailAlreadyExists,
		},
		"user might exist with this email, not sure due to sql mapping": {
			user: &model.User{},
			repo: &testUserRepo{
				errRetrieveUser:  resp.ErrUnknownUser,
				errRetrieveEmail: resp.ErrSQLMappingError,
			},
			expectedErr: resp.ErrCouldNotDetermineUserExistence,
		},
	}
	for tname, tcase := range tests {
		t.Run(tname, func(t *testing.T) {
			svc := NewUserService(cfg.TestConfig(), tcase.repo)
			beforeEncryptedPassword := tcase.user.Password
			err := svc.AddUser(tcase.user)

			if tcase.expectedErr != nil {
				assert.Equal(t, err.Error(), tcase.expectedErr.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tcase.repo.lastInsertId, tcase.user.Id)
				err := bcrypt.CompareHashAndPassword([]byte(tcase.user.Password), []byte(beforeEncryptedPassword))
				assert.Nil(t, err)
			}
		})
	}
}

func Test_RetrieveUnauthenticatedUser(t *testing.T) {
	tests := map[string]struct {
		userStr     string
		repo        *testUserRepo
		expectedErr error
	}{
		"valid email": {
			userStr: "benchadwick87@gmail.com",
			repo: &testUserRepo{
				userRetrieveUserByEmail: &model.User{
					Username: "ben",
					Email:    "benchadwick87@gmail.com",
					Password: "test-password",
					Id:       5,
				},
			},
		},
		"valid username": {
			userStr: "ben",
			repo: &testUserRepo{
				userRetrieveUserByUsername: &model.User{
					Username: "ben",
					Email:    "benchadwick87@gmail.com",
					Password: "test-password",
					Id:       5,
				},
			},
		},
		"invalid, could not determine user type": {
			userStr:     "ben!", // not an email, and it contains illegal character for username
			repo:        &testUserRepo{},
			expectedErr: resp.ErrCouldNotDetermineUserType,
		},
	}
	for tname, tcase := range tests {
		t.Run(tname, func(t *testing.T) {
			svc := NewUserService(cfg.TestConfig(), tcase.repo)
			user, err := svc.RetrieveUnauthenticatedUser(tcase.userStr)
			if tcase.expectedErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, tcase.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, user)
				assert.NotEmpty(t, user.Username)
				assert.NotEmpty(t, user.Email)
				assert.Empty(t, user.Password)
				assert.NotZero(t, user.Id)
			}
		})
	}
}

func Test_RetrieveAuthenticatedUser(t *testing.T) {
	tests := map[string]struct {
		userStr     string
		password    string
		repo        *testUserRepo
		expectedErr error
	}{
		"valid email": {
			userStr:  "benchadwick87@gmail.com",
			password: "Hello123!",
			repo: &testUserRepo{
				userRetrieveUserByEmail: &model.User{
					Username: "ben",
					Email:    "benchadwick87@gmail.com",
					Password: "$2a$10$QLgG8tbDrlpDUooY41Vz4elR173ckJexNqy/0eozaRwkURt6MEm3W",
					Id:       5,
				},
			},
		},
		"valid username": {
			userStr:  "ben",
			password: "Hello123!",
			repo: &testUserRepo{
				userRetrieveUserByUsername: &model.User{
					Username: "ben",
					Email:    "benchadwick87@gmail.com",
					Password: "$2a$10$QLgG8tbDrlpDUooY41Vz4elR173ckJexNqy/0eozaRwkURt6MEm3W",
					Id:       5,
				},
			},
		},
		"invalid, password was not right": {
			userStr:  "ben",
			password: "something other than Hello123!",
			repo: &testUserRepo{
				userRetrieveUserByUsername: &model.User{
					Username: "ben",
					Email:    "benchadwick87@gmail.com",
					Password: "$2a$10$QLgG8tbDrlpDUooY41Vz4elR173ckJexNqy/0eozaRwkURt6MEm3W",
					Id:       5,
				},
			},
			expectedErr: resp.ErrInvalidCredentials,
		},
		"invalid, could not determine user type": {
			userStr:     "ben!", // not an email, and it contains illegal character for username
			repo:        &testUserRepo{},
			expectedErr: resp.ErrCouldNotDetermineUserType,
		},
	}
	for tname, tcase := range tests {
		t.Run(tname, func(t *testing.T) {
			svc := NewUserService(cfg.TestConfig(), tcase.repo)
			user, err := svc.RetrieveAuthenticatedUser(tcase.userStr, tcase.password)
			if tcase.expectedErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, tcase.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, user)
				assert.NotEmpty(t, user.Username)
				assert.NotEmpty(t, user.Email)
				assert.Empty(t, user.Password)
				assert.NotZero(t, user.Id)
			}
		})
	}
}

type testUserRepo struct {
	errAddUser                 error
	errRetrieveEmail           error
	errRetrieveUser            error
	lastInsertId               int64
	userRetrieveUserByEmail    *model.User
	userRetrieveUserByUsername *model.User
}

func (trepo *testUserRepo) addUser(user *model.User) (int64, error) {
	return trepo.lastInsertId, trepo.errAddUser
}

func (trepo *testUserRepo) retrieveUserByEmail(userStr string) (*model.User, error) {
	return trepo.userRetrieveUserByEmail, trepo.errRetrieveEmail
}

func (trepo *testUserRepo) retrieveUserByUsername(userStr string) (*model.User, error) {
	return trepo.userRetrieveUserByUsername, trepo.errRetrieveUser
}
