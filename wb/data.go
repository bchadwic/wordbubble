package wb

import (
	"database/sql"
	"fmt"

	"github.com/bchadwic/wordbubble/model"
	"github.com/bchadwic/wordbubble/util"
	_ "github.com/mattn/go-sqlite3"
)

type DataSource interface {
	// users
	AddUser(user *model.User) (int64, error)
	RetrieveUserByString(userStr string) *model.User

	// wordbubbles
	AddNewWordBubble(userId int64, wb *model.WordBubble) error
	NumberOfWordBubblesForUser(userId int64) (int64, error)
	RemoveAndReturnLatestWordBubbleForUserId(userId int64) *model.WordBubble
}

type dataSource struct {
	db  *sql.DB
	log util.Logger
}

func NewDataSource(logger util.Logger) *dataSource {
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

func (source *dataSource) AddUser(user *model.User) (int64, error) {
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

func (source *dataSource) RetrieveUserByString(userStr string) *model.User {
	// TODO log the error once we get it back
	var stmt *sql.Stmt
	var err error
	switch {
	case util.ValidEmail(userStr) == nil:
		stmt, err = source.db.Prepare(`SELECT user_id, username, email, password FROM users WHERE email = ?`)
	case util.ValidUsername(userStr) == nil:
		stmt, err = source.db.Prepare(`SELECT user_id, username, email, password FROM users WHERE username = ?`)
	default:
		source.log.Info("couldn't determine if the string passed is a username or an email")
		return nil
	}
	if err != nil {
		source.log.Error("prepared statement error for retrieving user by string, user: %s, error: %s", userStr, err)
		return nil
	}
	row, err := stmt.Query(userStr)
	if err != nil {
		source.log.Error("querying error for retrieving user by string: %s, error: %s", userStr, err)
		return nil
	}
	defer row.Close()
	if !row.Next() {
		source.log.Info("could not find user in database, user: %s", userStr)
		return nil
	}
	var dbUser model.User
	if err := row.Scan(&dbUser.Id, &dbUser.Username, &dbUser.Email, &dbUser.Password); err != nil {
		source.log.Error("could not map db user to user struct, user: %s, error: %s", userStr, err)
		return nil
	}
	return &dbUser
}

func (source *dataSource) AddNewWordBubble(userId int64, wb *model.WordBubble) error {
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

func (source *dataSource) RemoveAndReturnLatestWordBubbleForUserId(userId int64) *model.WordBubble {
	source.log.Info("removing and returning the last wordbubble for %d", userId)
	stmt, err := source.db.Prepare(`
	DELETE FROM wordbubbles WHERE wordbubble_id = ( 
		SELECT wordbubble_id FROM wordbubbles WHERE user_id = ? ORDER BY created_timestamp ASC LIMIT 1
	) RETURNING text;`)
	if err != nil {
		source.log.Error("prepared statement error for removing and returning wordbubble for user: %d, error: %s", userId, err)
		return nil
	}
	row, err := stmt.Query(userId)
	if err != nil {
		source.log.Error("querying error for removing and returning wordbubble for user: %d, error: %s", userId, err)
		return nil
	}
	defer row.Close()
	if !row.Next() {
		source.log.Error("no wordbubble to return for user %d", userId)
		return nil
	}
	var wordbubble model.WordBubble
	if err := row.Scan(&wordbubble.Text); err != nil {
		source.log.Error("could not map db wordbubble text for user: %s, error: %s", userId, err)
		return nil
	}
	return &wordbubble
}
