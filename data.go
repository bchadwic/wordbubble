package main

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type DataSource interface {
	// users
	AddUser(user *User) (int64, error)
	GetAuthenticatedUserFromUsername(user *User) (*User, error)
	GetUserFromUsername(username string) (*User, error)
	GetUserFromEmail(email string) (*User, error)
	ResolveUserIdFromUsername(email string) (int64, error)
	ResolveUserIdFromEmail(email string) (int64, error)
	// wordbubbles
	AddNewWordBubble(userId int64, wb *WordBubble) error
	NumberOfWordBubblesForUser(userId int64) (int64, error)
	RemoveAndReturnLatestWordBubbleForUser(userId int64) (*WordBubble, error)
}

type dataSource struct {
	db  *sql.DB
	log Logger
}

func NewDataSource(logger Logger) *dataSource {
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
	return &dataSource{db, logger}
}

func (source *dataSource) AddUser(user *User) (int64, error) {
	stmt, err := source.db.Prepare(`INSERT INTO users(username, email, password) VALUES (?, ?, ?);`)
	if err != nil {
		source.log.Error("prepared statement error for adding user: %s, error: %s", user.Username, err)
		return 0, err
	}
	res, err := stmt.Exec(user.Username, user.Email, user.Password)
	if err != nil {
		source.log.Error("executing error for adding user: %s, error: %s", user.Username, err)
		return 0, err
	}
	return res.LastInsertId()
}

func (source *dataSource) GetAuthenticatedUserFromUsername(user *User) (*User, error) {
	stmt, err := source.db.Prepare(`SELECT user_id, username, email, password FROM users WHERE username = ?`)
	if err != nil {
		source.log.Error("prepared statement error for retrieving authenticated user: %s, error: %s", user.Username, err)
		return nil, err
	}
	row, err := stmt.Query(user.Username)
	if err != nil {
		source.log.Error("querying error for retrieving authenticated user: %s, error: %s", user.Username, err)
		return nil, err
	}
	defer row.Close()
	if !row.Next() {
		source.log.Error("could not find user in database, user: %s", user.Username)
		return nil, fmt.Errorf("could not find user with username %s", user.Username)
	}
	var dbUser User
	if err := row.Scan(&dbUser.UserId, &dbUser.Username, &dbUser.Email, &dbUser.Password); err != nil {
		source.log.Error("could not map db user to user struct, user: %s, error: %s", user.Username, err)
		return nil, fmt.Errorf("could not retrive user information for %s", user.Username)
	}
	return &dbUser, nil
}

func (source *dataSource) GetUserFromUsername(username string) (*User, error) {
	stmt, err := source.db.Prepare(`SELECT user_id, username, email FROM users WHERE username = ?`)
	if err != nil {
		source.log.Error("prepared statement error for getting user from username: %s, error: %s", username, err)
		return nil, err
	}
	row, err := stmt.Query(username)
	if err != nil {
		source.log.Error("querying error for getting user from username: %s, error: %s", username, err)
		return nil, err
	}
	defer row.Close()
	if !row.Next() {
		source.log.Error("could not find a user with username: %s", username)
		return nil, fmt.Errorf("could not find user with username %s", username)
	}
	var user User
	row.Scan(&user.UserId, &user.Username, &user.Email)
	return &user, nil
}

func (source *dataSource) GetUserFromEmail(email string) (*User, error) {
	source.log.Info("retrieving user by email %s", email)
	stmt, err := source.db.Prepare(`SELECT user_id, username, email FROM users WHERE email = ?`)
	if err != nil {
		source.log.Error("prepared statement error for getting user from email: %s, error: %s", email, err)
		return nil, err
	}
	row, err := stmt.Query(email)
	if err != nil {
		source.log.Error("query error for getting user from email: %s, error: %s", email, err)
		return nil, err
	}
	defer row.Close()
	if !row.Next() {
		source.log.Error("could not find user with email: %s", email)
		return nil, fmt.Errorf("could not find user with email %s", email)
	}
	var user User
	row.Scan(&user.UserId, &user.Username, &user.Email)
	return &user, nil
}

func (source *dataSource) ResolveUserIdFromUsername(username string) (int64, error) {
	source.log.Info("retrieving user id for %s", username)
	stmt, err := source.db.Prepare(`SELECT user_id FROM users WHERE username = ?`)
	if err != nil {
		source.log.Error("prepared statement error for getting userId from username: %s, error: %s", username, err)
		return 0, err
	}
	row, err := stmt.Query(username)
	if err != nil {
		source.log.Error("querying error for getting userId from username: %s, error: %s", username, err)
		return 0, err
	}
	defer row.Close()
	if !row.Next() {
		source.log.Error("could not find user with username: %s", username)
		return 0, fmt.Errorf("could not find %s", username)
	}
	var userId int64
	if err := row.Scan(&userId); err != nil {
		source.log.Error("could not map db userId for user: %s, error: %s", username, err)
		return 0, fmt.Errorf("could not parse identity for %s", username)
	}
	return userId, nil
}

func (source *dataSource) ResolveUserIdFromEmail(email string) (int64, error) {
	stmt, err := source.db.Prepare(`SELECT user_id FROM users WHERE email = ?`)
	if err != nil {
		source.log.Error("prepared statement error for getting userId from email: %s, error: %s", email, err)
		return 0, err
	}
	row, err := stmt.Query(email)
	if err != nil {
		source.log.Error("querying error for getting userId from email: %s, error: %s", email, err)
		return 0, err
	}
	defer row.Close()
	if !row.Next() {
		source.log.Error("could not find user with email: %s", email)
		return 0, fmt.Errorf("could not find %s", email)
	}
	var userId int64
	if err := row.Scan(&userId); err != nil {
		source.log.Error("could not map db userId for user: %s, error: %s", email, err)
		return 0, fmt.Errorf("could not parse identity for %s", email)
	}
	return userId, nil
}

func (source *dataSource) AddNewWordBubble(userId int64, wb *WordBubble) error {
	stmt, err := source.db.Prepare(`INSERT INTO wordbubbles (user_id, text) VALUES (?, ?)`)
	if err != nil {
		source.log.Error("prepared statement error for adding a wordbubble %+v for user: %d, error: %s", wb, userId, err)
		return err
	}
	_, err = stmt.Exec(userId, wb.Text)
	if err != nil {
		source.log.Error("execute error for adding a wordbubble %+v for user: %d, error: %s", wb, userId, err)
		return err
	}
	return nil
}

func (source *dataSource) NumberOfWordBubblesForUser(userId int64) (int64, error) {
	source.log.Info("checking amount of wordbubbles for user %d", userId)
	stmt, err := source.db.Prepare(`SELECT COUNT(*) from wordbubbles WHERE user_id = ?`)
	if err != nil {
		source.log.Error("prepared statement error for getting number of wordbubbles for user: %d, error: %s", userId, err)
		return 0, err
	}
	row, err := stmt.Query(userId)
	if err != nil {
		source.log.Error("query error for getting number of wordbubbles for user: %d, error: %s", userId, err)
		return 0, err
	}
	defer row.Close()
	if !row.Next() {
		source.log.Error("could not find how many wordbubbles for user: %d", userId)
		return 0, fmt.Errorf("an error occurred gathering the current amount of wordbubbles")
	}
	var amt int64
	if err := row.Scan(&amt); err != nil {
		source.log.Error("could not map db wordbubble amount for user: %s, error: %s", userId, err)
		return 0, fmt.Errorf("an error occurred gathering the current amount of wordbubbles for user")
	}
	return amt, nil
}

func (source *dataSource) RemoveAndReturnLatestWordBubbleForUser(userId int64) (*WordBubble, error) {
	source.log.Info("removing and returning the last wordbubble for %d", userId)
	stmt, err := source.db.Prepare(`
	DELETE FROM wordbubbles WHERE wordbubble_id = ( 
		SELECT wordbubble_id FROM wordbubbles WHERE user_id = ? ORDER BY created_timestamp ASC LIMIT 1
	) RETURNING text;`)
	if err != nil {
		source.log.Error("prepared statement error for removing and returning wordbubble for user: %d, error: %s", userId, err)
		return nil, err
	}
	row, err := stmt.Query(userId)
	if err != nil {
		source.log.Error("querying error for removing and returning wordbubble for user: %d, error: %s", userId, err)
		return nil, err
	}
	defer row.Close()
	if !row.Next() {
		source.log.Error("no wordbubble to return for user %d", userId)
		return nil, nil // TODO make this better
	}
	var wordbubble WordBubble
	if err := row.Scan(&wordbubble.Text); err != nil {
		source.log.Error("could not map db wordbubble text for user: %s, error: %s", userId, err)
		return nil, errors.New("an error occurred after removing wordbubble from database")
	}
	return &wordbubble, nil
}
