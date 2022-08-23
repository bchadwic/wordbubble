package main

import "fmt"

type DB interface {
	AddUser(logger Logger, user *User) error
	GetUserFromUsername(logger Logger, username string) (*User, error)
}

type db struct {
	users map[string]*User
}

func NewDB() *db {
	return &db{
		users: make(map[string]*User, 0),
	}
}

func (db *db) AddUser(logger Logger, user *User) error {
	logger.Info("db.AddUser: adding in new user %s", user.Username)
	db.users[user.Username] = user
	logger.Info("db.AddUser: successfully user %s added to the database", user.Username)
	return nil
}

func (db *db) GetUserFromUsername(logger Logger, username string) (*User, error) {
	logger.Info("db.GetUserFromUsername: retrieving user %s", username)
	user, ok := db.users[username]
	if !ok {
		logger.Error("db.GetUserFromUsername: could not find user with username %s", username)
		return nil, fmt.Errorf("could not find user")
	}
	logger.Info("db.GetUserFromUsername: successfully found %s in the database", username)
	return user, nil
}
