package dao

import (
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Word struct {
	ID       int `gorm:"primaryKey;autoIncrement"`
	Word     string
	Kana     string
	Username string
	UserID   int64
}

type Player struct {
	ID        int64  `gorm:"primaryKey"`
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

func (dbc *DBConnection) Init() {
	dbc.Error = dbc.doWithRetryConnection(func(db *gorm.DB) error {
		db.Migrator().DropTable(&Word{}, &Player{})
		return db.AutoMigrate(&Word{}, &Player{})
	})
}

func (dbc *DBConnection) ClearTables() {
	dbc.Error = dbc.doWithRetryConnection(func(db *gorm.DB) error {
		var tables []string
		if err := db.Raw(`SELECT tablename FROM pg_tables WHERE schemaname = 'public'`).Scan(&tables).Error; err != nil {
			return err
		}
		for _, t := range tables {
			if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", t)).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (dbc *DBConnection) AddPlayer(id int64, username, firstName string) error {
	return dbc.doWithRetryConnection(func(db *gorm.DB) error {
		return db.Create(&Player{ID: id, Username: username, FirstName: firstName, Score: 0}).Error
	})
}

func (dbc *DBConnection) AllPlayers() []Player {
	var players []Player
	dbc.Error = dbc.doWithRetryConnection(func(db *gorm.DB) error {
		return db.Find(&players).Error
	})

	return players
}

func (dbc *DBConnection) CheckPlayerExistence(username string) bool {
	var count int64
	dbc.Error = dbc.doWithRetryConnection(func(db *gorm.DB) error {
		return db.Model(&Player{}).Where("username = ?", username).Count(&count).Error
	})
	return count > 0
}

func (dbc *DBConnection) AddWord(word, kana, username string, userID int64) {
	dbc.Error = dbc.doWithRetryConnection(func(db *gorm.DB) error {
		if err := db.Create(&Word{UserID: userID, Username: username, Word: word, Kana: kana}).Error; err != nil {
			return err
		}
		return db.Model(&Player{}).Where("username = ?", username).
			Update("score", gorm.Expr("score + 1")).Error
	})
}

func (dbc *DBConnection) SetScore(username string, score uint64) {
	dbc.Error = dbc.doWithRetryConnection(func(db *gorm.DB) error {
		return db.Model(&Player{}).Where("username = ?", username).
			Update("score", score).Error
	})
}

func (dbc *DBConnection) LastWord() (string, string) {
	var last Word
	dbc.Error = dbc.doWithRetryConnection(func(db *gorm.DB) error {
		return db.Last(&last).Error
	})
	return last.Word, last.Kana
}

func (dbc *DBConnection) LastPlayer() int64 {
	var last Word
	dbc.Error = dbc.doWithRetryConnection(func(db *gorm.DB) error {
		return db.Last(&last).Error
	})

	return last.UserID
}

func (dbc *DBConnection) CheckWordExistence(word string) bool {
	var count int64 = 0
	dbc.Error = dbc.doWithRetryConnection(func(db *gorm.DB) error {
		return db.Model(&Word{}).Where("word = ?", word).Count(&count).Error
	})

	return count > 0
}

func (dbc *DBConnection) Reset() {
	dbc.Error = nil
}
