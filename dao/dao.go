package dao

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type GameKey struct {
	ChatID   int64 `gorm:"column:chat_id;not null;index"`
	ThreadID int   `gorm:"column:thread_id;not null;index"`
}

type Word struct {
	ID       int `gorm:"primaryKey;autoIncrement"`
	GameKey  `gorm:"embedded"`
	Word     string
	Kana     string
	Username string
	UserID   int64
}

type Player struct {
	ID        int64 `gorm:"primaryKey"`
	GameKey   `gorm:"embedded"`
	FirstName string //Pretty stats at the end
	Username  string
	Score     uint64
}

const (
	maxRetries          = 3
	delayBetweenRetries = 2 * time.Second
)

func connectToDatabase(dsn string) (*gorm.DB, error) {
	var dbConn *gorm.DB
	var err error
	for i := 1; i <= maxRetries; i++ {
		dbConn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			log.Println("Connect to PostgreSQL")
			return dbConn, nil
		}
		log.Printf("Couldn't connect to PostgreSQL (attempt %d/%d): %v\nRetrying in %v...", i, maxRetries, err, delayBetweenRetries)
		time.Sleep(delayBetweenRetries)
	}
	return nil, err
}

type DBConnection struct {
	dsn    string
	dbConn *gorm.DB
	Error  error
}

func NewConnection(dsn string) (*DBConnection, error) {
	dbConn, err := connectToDatabase(dsn)
	if err != nil {
		return nil, err
	}
	return &DBConnection{dsn: dsn, dbConn: dbConn}, nil
}

func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "connection already closed") ||
		strings.Contains(msg, "database is closed") ||
		strings.Contains(msg, "eof") ||
		strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "network is unreachable")
}

func (dbc *DBConnection) doWithRetryConnection(fn func(*gorm.DB) error) error {
	err := fn(dbc.dbConn)
	if err == nil {
		return nil
	}

	if isConnectionError(err) {
		log.Println("Lost DB connection, reconnecting...")
		newDB, connErr := connectToDatabase(dbc.dsn)
		if connErr != nil {
			return fmt.Errorf("reconnect failed: %w", connErr)
		}
		dbc.dbConn = newDB
		err = fn(dbc.dbConn)
	}

	return err
}

func (dbc *DBConnection) Init(key GameKey) {
	dbc.Error = dbc.doWithRetryConnection(func(db *gorm.DB) error {
		return db.AutoMigrate(&Word{}, &Player{})
	})
}

func (dbc *DBConnection) ClearGame(key GameKey) {
	dbc.Error = dbc.doWithRetryConnection(func(db *gorm.DB) error {
		if err := db.
			Where("chat_id = ? AND thread_id = ?", key.ChatID, key.ThreadID).
			Delete(&Word{}).Error; err != nil {
			return err
		}

		if err := db.
			Where("chat_id = ? AND thread_id = ?", key.ChatID, key.ThreadID).
			Delete(&Player{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (dbc *DBConnection) AddPlayer(key GameKey, id int64, username, firstName string) error {
	return dbc.doWithRetryConnection(func(db *gorm.DB) error {
		return db.Create(&Player{
			ID:        id,
			GameKey:   key,
			Username:  username,
			FirstName: firstName,
			Score:     0,
		}).Error
	})
}

func (dbc *DBConnection) AllPlayers(key GameKey) []Player {
	var players []Player
	dbc.Error = dbc.doWithRetryConnection(func(db *gorm.DB) error {
		return db.
			Where("chat_id = ? AND thread_id = ?", key.ChatID, key.ThreadID).
			Order("score DESC").
			Find(&players).Error
	})

	return players
}

func (dbc *DBConnection) CheckPlayerExistence(key GameKey, username string) bool {
	var count int64
	dbc.Error = dbc.doWithRetryConnection(func(db *gorm.DB) error {
		return db.Model(&Player{}).
			Where(
				"chat_id = ? AND thread_id = ? AND username = ?",
				key.ChatID, key.ThreadID, username,
			).
			Count(&count).Error
	})

	return count > 0
}

func (dbc *DBConnection) AddWord(key GameKey, word, kana, username string, userID int64) {
	dbc.Error = dbc.doWithRetryConnection(func(db *gorm.DB) error {

		if err := db.Create(&Word{
			GameKey:  key,
			Word:     word,
			Kana:     kana,
			Username: username,
			UserID:   userID,
		}).Error; err != nil {
			return err
		}

		return db.Model(&Player{}).
			Where(
				"chat_id = ? AND thread_id = ? AND username = ?",
				key.ChatID, key.ThreadID, username,
			).
			Update("score", gorm.Expr("score + 1")).Error
	})
}

func (dbc *DBConnection) SetScore(key GameKey, username string, score uint64) {
	dbc.Error = dbc.doWithRetryConnection(func(db *gorm.DB) error {
		return db.Model(&Player{}).
			Where(
				"chat_id = ? AND thread_id = ? AND username = ?",
				key.ChatID, key.ThreadID, username,
			).
			Update("score", score).Error
	})
}

func (dbc *DBConnection) LastWord(key GameKey) (string, string) {
	var last Word
	dbc.Error = dbc.doWithRetryConnection(func(db *gorm.DB) error {
		err := db.
			Where(
				"chat_id = ? AND thread_id = ?",
				key.ChatID, key.ThreadID,
			).
			Order("id DESC").
			First(&last).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil // нет слов — игра ещё не началась
			}
			return err
		}

		return nil
	})

	return last.Word, last.Kana
}

func (dbc *DBConnection) LastPlayer(key GameKey) int64 {
	var last Word
	dbc.Error = dbc.doWithRetryConnection(func(db *gorm.DB) error {
		return db.
			Where(
				"chat_id = ? AND thread_id = ?",
				key.ChatID, key.ThreadID,
			).
			Order("id DESC").
			First(&last).Error
	})

	return last.UserID
}

func (dbc *DBConnection) CheckWordExistence(key GameKey, word string) bool {
	var count int64
	dbc.Error = dbc.doWithRetryConnection(func(db *gorm.DB) error {
		return db.Model(&Word{}).
			Where(
				"chat_id = ? AND thread_id = ? AND word = ?",
				key.ChatID, key.ThreadID, word,
			).
			Count(&count).Error
	})

	return count > 0
}

func (dbc *DBConnection) Reset() {
	dbc.Error = nil
}
