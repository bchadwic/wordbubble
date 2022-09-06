package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type DataSource interface {
	// users
	AddUser(logger Logger, user *User) (int64, error)
	GetAuthenticatedUserFromUsername(logger Logger, user *User) (*User, error)
	GetUserFromUsername(logger Logger, username string) (*User, error)
	GetUserFromEmail(logger Logger, email string) (*User, error)
	ResolveUserIdFromUsername(logger Logger, email string) (int64, error)
	ResolveUserIdFromEmail(logger Logger, email string) (int64, error)
	// wordbubbles
	AddNewWordBubble(logger Logger, userId int64, wb *WordBubble) error
	NumberOfWordBubblesForUser(logger Logger, userId int64) (int64, error)
	RemoveAndReturnLatestWordBubbleForUser(logger Logger, userId int64) (*WordBubble, error)
}

type datasource struct {
	db *sql.DB
}

func NewDataSource() *datasource {
	panicker := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	db, err := sql.Open("sqlite3", "./wb-database.db")
	panicker(err)
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
	);`)
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
	return &datasource{db}
}

func (ds *datasource) AddUser(logger Logger, user *User) (int64, error) {
	logger.Info("db.AddUser: adding in new user %s", user.Username)
	stmt, err := ds.db.Prepare(`INSERT INTO users(username, email, password) VALUES (?, ?, ?);`)
	if err != nil {
		return 0, err
	}
	res, err := stmt.Exec(user.Username, user.Email, user.Password)
	if err != nil {
		return 0, err
	}
	logger.Info("db.AddUser: successfully user %s added to the database", user.Username)
	return res.LastInsertId()
}

func (ds *datasource) GetAuthenticatedUserFromUsername(logger Logger, user *User) (*User, error) {
	logger.Info("db.GetAuthenticatedUserFromUsername: retrieving user %s", user.Username)
	stmt, err := ds.db.Prepare(`SELECT user_id, username, email, password FROM users WHERE username = ?`)
	if err != nil {
		return nil, err
	}
	row, err := stmt.Query(user.Username)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	if !row.Next() {
		return nil, fmt.Errorf("could not find user with username %s", user.Username)
	}
	var dbUser User
	if err := row.Scan(&dbUser.UserId, &dbUser.Username, &dbUser.Email, &dbUser.Password); err != nil {
		return nil, fmt.Errorf("could not retrive user information for %s", user.Username)
	}
	logger.Info("db.GetUserFromUsername: successfully found %s in the database", dbUser.Username)
	return &dbUser, nil
}

func (ds *datasource) GetUserFromUsername(logger Logger, username string) (*User, error) {
	logger.Info("db.GetUserFromUsername: retrieving user %s", username)
	stmt, err := ds.db.Prepare(`SELECT user_id, username, email FROM users WHERE username = ?`)
	if err != nil {
		return nil, err
	}
	row, err := stmt.Query(username)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	if !row.Next() {
		return nil, fmt.Errorf("could not find user with username %s", username)
	}
	var user User
	row.Scan(&user.UserId, &user.Username, &user.Email)
	logger.Info("db.GetUserFromUsername: successfully found %s in the database", username)
	return &user, nil
}

func (ds *datasource) GetUserFromEmail(logger Logger, email string) (*User, error) {
	logger.Info("db.GetUserFromEmail: retrieving user by email %s", email)
	stmt, err := ds.db.Prepare(`SELECT user_id, username, email FROM users WHERE email = ?`)
	if err != nil {
		return nil, err
	}
	row, err := stmt.Query(email)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	if !row.Next() {
		return nil, fmt.Errorf("could not find user with email %s", email)
	}
	var user User
	row.Scan(&user.UserId, &user.Username, &user.Email)
	logger.Info("db.GetUserFromUsername: successfully found %s in the database", user.Username)
	return &user, nil
}

func (ds *datasource) ResolveUserIdFromUsername(logger Logger, username string) (int64, error) {
	logger.Info("db.ResolveUserIdFromUsername: retrieving user id for %s", username)
	stmt, err := ds.db.Prepare(`SELECT user_id FROM users WHERE username = ?`)
	if err != nil {
		return 0, err
	}
	row, err := stmt.Query(username)
	if err != nil {
		return 0, err
	}
	defer row.Close()
	if !row.Next() {
		return 0, fmt.Errorf("could not find %s", username)
	}
	var userId int64
	if err := row.Scan(&userId); err != nil {
		return 0, fmt.Errorf("could not parse identity for %s", username)
	}
	logger.Info("db.ResolveUserIdFromUsername: successfully found userId: %d in the database", userId)
	return userId, nil
}

func (ds *datasource) ResolveUserIdFromEmail(logger Logger, email string) (int64, error) {
	logger.Info("db.ResolveUserIdFromEmail: retrieving user id for %s", email)
	stmt, err := ds.db.Prepare(`SELECT user_id FROM users WHERE email = ?`)
	if err != nil {
		return 0, err
	}
	row, err := stmt.Query(email)
	if err != nil {
		return 0, err
	}
	defer row.Close()
	if !row.Next() {
		return 0, fmt.Errorf("could not find %s", email)
	}
	var userId int64
	if err := row.Scan(&userId); err != nil {
		return 0, fmt.Errorf("could not parse identity for %s", email)
	}
	logger.Info("db.ResolveUserIdFromEmail: successfully found userId: %d in the database", userId)
	return userId, nil
}

func (ds *datasource) AddNewWordBubble(logger Logger, userId int64, wb *WordBubble) error {
	logger.Info("db.AddNewWordBubble: creating wordbubble %+v for %d", wb, userId)
	logger.Info("db.AddNewWordBubble: %+v ", ds.db.Stats())
	stmt, err := ds.db.Prepare(`INSERT INTO wordbubbles (user_id, text) VALUES (?, ?)`)
	if err != nil {
		logger.Error("db.AddNewWordBubble: could not prepare statement: %s", err)
		return err
	}
	_, err = stmt.Exec(userId, wb.Text)
	if err != nil {
		logger.Error("db.AddNewWordBubble: could not execute statement: %s", err)
		return err
	}
	logger.Info("db.AddNewWordBubble: successfully added a new wordbubble for user %d", userId)
	return nil
}

func (ds *datasource) NumberOfWordBubblesForUser(logger Logger, userId int64) (int64, error) {
	logger.Info("db.NumberOfWordBubblesForUser: checking amount of wordbubbles for user %d", userId)
	stmt, err := ds.db.Prepare(`SELECT COUNT(*) from wordbubbles WHERE user_id = ?`)
	if err != nil {
		logger.Error("db.AddNewWordBubble: could not prepare statement: %s", err)
		return 0, err
	}
	row, err := stmt.Query(userId)
	if err != nil {
		logger.Error("db.AddNewWordBubble: could not query using statement: %s", err)
		return 0, err
	}
	defer row.Close()
	if !row.Next() {
		logger.Error("db.AddNewWordBubble: could not find how many wordbubbles for user: %d", userId)
		return 0, fmt.Errorf("an error occurred gathering the current amount of wordbubbles")
	}
	var amt int64
	if err := row.Scan(&amt); err != nil {
		logger.Error("db.AddNewWordBubble: could not find how many wordbubbles for user: %d", userId)
		return 0, fmt.Errorf("an error occurred gathering the current amount of wordbubbles")
	}
	logger.Info("db.AddNewWordBubble: successfully added a new wordbubble for user %d", userId)
	return amt, nil
}

func (ds *datasource) RemoveAndReturnLatestWordBubbleForUser(logger Logger, userId int64) (*WordBubble, error) {
	logger.Info("db.RemoveAndReturnLatestWordBubbleForUser: removing and returning the last wordbubble for %d", userId)
	stmt, err := ds.db.Prepare(`
	DELETE FROM wordbubbles WHERE wordbubble_id = ( 
		SELECT wordbubble_id FROM wordbubbles WHERE user_id = ? ORDER BY created_timestamp DESC LIMIT 1
	) RETURNING text;
	`)
	if err != nil {
		logger.Error("db.RemoveAndReturnLatestWordBubbleForUser: could not prepare statement: %s", err)
		return nil, err
	}
	row, err := stmt.Query(userId)
	if err != nil {
		logger.Error("db.RemoveAndReturnLatestWordBubbleForUser: could not query using statement: %s", err)
		return nil, err
	}
	defer row.Close()
	if !row.Next() {
		logger.Warn("db.RemoveAndReturnLatestWordBubbleForUser: user %d did not have a wordbubble to return", userId)
		return nil, nil // TODO make this better? no wordbubble found
	}
	var wordbubble WordBubble
	if err := row.Scan(&wordbubble.Text); err != nil {
		logger.Error("db.RemoveAndReturnLatestWordBubbleForUser: unable to parse row from result set")
		return nil, err
	}
	logger.Info("db.RemoveAndReturnLatestWordBubbleForUser: successfully removed a wordbubble for user %d, now returning", userId)
	return &wordbubble, nil
}
