package main

import (
	"database/sql"
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
	source.log.Info("adding in new user %s", user.Username)
	stmt, err := source.db.Prepare(`INSERT INTO users(username, email, password) VALUES (?, ?, ?);`)
	if err != nil {
		return 0, err
	}
	res, err := stmt.Exec(user.Username, user.Email, user.Password)
	if err != nil {
		return 0, err
	}
	source.log.Info("successfully user %s added to the database", user.Username)
	return res.LastInsertId()
}

func (source *dataSource) GetAuthenticatedUserFromUsername(user *User) (*User, error) {
	source.log.Info("retrieving user %s", user.Username)
	stmt, err := source.db.Prepare(`SELECT user_id, username, email, password FROM users WHERE username = ?`)
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
	source.log.Info("successfully found %s in the database", dbUser.Username)
	return &dbUser, nil
}

func (source *dataSource) GetUserFromUsername(username string) (*User, error) {
	source.log.Info("retrieving user %s", username)
	stmt, err := source.db.Prepare(`SELECT user_id, username, email FROM users WHERE username = ?`)
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
	source.log.Info("successfully found %s in the database", username)
	return &user, nil
}

func (source *dataSource) GetUserFromEmail(email string) (*User, error) {
	source.log.Info("retrieving user by email %s", email)
	stmt, err := source.db.Prepare(`SELECT user_id, username, email FROM users WHERE email = ?`)
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
	source.log.Info("successfully found %s in the database", user.Username)
	return &user, nil
}

func (source *dataSource) ResolveUserIdFromUsername(username string) (int64, error) {
	source.log.Info("retrieving user id for %s", username)
	stmt, err := source.db.Prepare(`SELECT user_id FROM users WHERE username = ?`)
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
	source.log.Info("successfully found userId: %d in the database", userId)
	return userId, nil
}

func (source *dataSource) ResolveUserIdFromEmail(email string) (int64, error) {
	source.log.Info("retrieving user id for %s", email)
	stmt, err := source.db.Prepare(`SELECT user_id FROM users WHERE email = ?`)
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
	source.log.Info("successfully found userId: %d in the database", userId)
	return userId, nil
}

func (source *dataSource) AddNewWordBubble(userId int64, wb *WordBubble) error {
	source.log.Info("creating wordbubble %+v for %d", wb, userId)
	stmt, err := source.db.Prepare(`INSERT INTO wordbubbles (user_id, text) VALUES (?, ?)`)
	if err != nil {
		source.log.Error("could not prepare statement: %s", err)
		return err
	}
	_, err = stmt.Exec(userId, wb.Text)
	if err != nil {
		source.log.Error("could not execute statement: %s", err)
		return err
	}
	source.log.Info("successfully added a new wordbubble for user %d", userId)
	return nil
}

func (source *dataSource) NumberOfWordBubblesForUser(userId int64) (int64, error) {
	source.log.Info("checking amount of wordbubbles for user %d", userId)
	stmt, err := source.db.Prepare(`SELECT COUNT(*) from wordbubbles WHERE user_id = ?`)
	if err != nil {
		source.log.Error("could not prepare statement: %s", err)
		return 0, err
	}
	row, err := stmt.Query(userId)
	if err != nil {
		source.log.Error("could not query using statement: %s", err)
		return 0, err
	}
	defer row.Close()
	if !row.Next() {
		source.log.Error("could not find how many wordbubbles for user: %d", userId)
		return 0, fmt.Errorf("an error occurred gathering the current amount of wordbubbles")
	}
	var amt int64
	if err := row.Scan(&amt); err != nil {
		source.log.Error("could not find how many wordbubbles for user: %d", userId)
		return 0, fmt.Errorf("an error occurred gathering the current amount of wordbubbles")
	}
	source.log.Info("successfully added a new wordbubble for user %d", userId)
	return amt, nil
}

func (source *dataSource) RemoveAndReturnLatestWordBubbleForUser(userId int64) (*WordBubble, error) {
	source.log.Info("removing and returning the last wordbubble for %d", userId)
	stmt, err := source.db.Prepare(`
	DELETE FROM wordbubbles WHERE wordbubble_id = ( 
		SELECT wordbubble_id FROM wordbubbles WHERE user_id = ? ORDER BY created_timestamp ASC LIMIT 1
	) RETURNING text;
	`)
	if err != nil {
		source.log.Error("could not prepare statement: %s", err)
		return nil, err
	}
	row, err := stmt.Query(userId)
	if err != nil {
		source.log.Error("could not query using statement: %s", err)
		return nil, err
	}
	defer row.Close()
	if !row.Next() {
		source.log.Warn("user %d did not have a wordbubble to return", userId)
		return nil, nil // TODO make this better? no wordbubble found
	}
	var wordbubble WordBubble
	if err := row.Scan(&wordbubble.Text); err != nil {
		source.log.Error("unable to parse row from result set")
		return nil, err
	}
	source.log.Info("successfully removed a wordbubble for user %d, now returning", userId)
	return &wordbubble, nil
}
