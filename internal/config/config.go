package config

import (
	"database/sql"
	"log"
	"os"

	"github.com/bchadwic/wordbubble/util"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Config interface {
	NewLogger(namespace string) util.Logger
	DB() *sql.DB
	Port() string
	Timer() util.Timer
}

type config struct {
	db *sql.DB
}

type testConfig struct {
	db    *sql.DB
	timer util.Timer
}

// NewConfig sets the configuration for the api using the environment settings
// the config returned may be nil if not all the dependencies could be successfully created
func NewConfig() *config {
	var cfg config
	log := cfg.NewLogger("config")
	log.Info("generating configuration")
	util.SigningKey = func() []byte {
		if s := os.Getenv("WB_SIGNING_KEY"); s != "" {
			return []byte(s)
		}
		return nil
	}
	if util.SigningKey() == nil {
		log.Error("signing key is not set")
		return nil
	}
	db, err := sql.Open("postgres", os.Getenv("DSN"))
	if err != nil {
		log.Error("db creation failed: " + err.Error())
		return nil
	}
	if err := db.Ping(); err != nil {
		log.Error("db ping failed: " + err.Error())
		return nil
	}
	cfg.db = db
	return &cfg
}

// TestConfig is used for unit testing only, do not use for any other scenario
func TestConfig() *testConfig {
	util.SigningKey = func() []byte { return []byte("test key") }
	var cfg testConfig
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
		return nil
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			user_id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS wordbubbles (
			wordbubble_id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,  
			created_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			text TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS tokens (
			user_id INTEGER NOT NULL,  
			refresh_token TEXT NOT NULL,
			issued_at INTEGER NOT NULL
		);
	`)
	if err != nil {
		log.Fatal(err)
	}
	cfg.db = db
	return &cfg
}

func (cfg *config) NewLogger(namespace string) util.Logger {
	return util.NewLogger(namespace, os.Getenv("WB_LOG_LEVEL"))
}

func (cfg *config) DB() *sql.DB {
	return cfg.db
}

func (cfg *config) Port() string {
	if p := os.Getenv("WB_PORT"); p != "" {
		return p
	}
	return ":8080"
}

func (cfg *config) Timer() util.Timer {
	return util.NewTimer()
}

func (cfg *testConfig) NewLogger(namespace string) util.Logger {
	return util.TestLogger()
}

func (cfg *testConfig) DB() *sql.DB {
	return cfg.db
}

func (cfg *testConfig) Port() string {
	return ":8080"
}

func (cfg *testConfig) Timer() util.Timer {
	return cfg.timer
}

func (cfg *testConfig) SetTimer(timer util.Timer) {
	cfg.timer = timer
}
