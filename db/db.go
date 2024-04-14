package db

import (
	"gorm.io/gorm"
)

type Word struct {
	ID       int `gorm:"primaryKey;autoIncrement"`
	Word     string
	Username string
}

type Player struct {
	ID       int `gorm:"primaryKey;autoIncrement"`
	Username string
	Score    uint64
}

func Init(db *gorm.DB) {
	if db.Migrator().HasTable(&Word{}) {
		db.Migrator().DropTable(&Word{})
	}
	if db.Migrator().HasTable(&Player{}) {
		db.Migrator().DropTable(&Player{})
	}
	db.AutoMigrate(&Word{})
	db.AutoMigrate(&Player{})
}

func ShutDown(db *gorm.DB) {
	db.Migrator().DropTable(&Player{}, &Word{})
}

func AddPlayer(db *gorm.DB, username string) {
	db.Create(&Player{Username: username, Score: 0})
}

func CheckPlayerExistence(db *gorm.DB, username string) bool {
	var players []Player
	db.Model(&Player{}).Where("username = ?", username).Find(&players)
	return len(players) != 0
}

func AddWord(db *gorm.DB, word string, from string) {
	db.Create(&Word{Username: from, Word: word})
	db.Model(&Player{}).Where("username = ?", from).Update("score", gorm.Expr("score + ?", 1))
}

func GetLastWord(db *gorm.DB) string {
	var lastWord Word
	db.Last(&lastWord)
	return lastWord.Word
}

func CheckWordExistence(db *gorm.DB, word string) bool {
	var words []Word
	db.Model(&Word{}).Where("word = ?", word).Find(&words)
	return len(words) != 0
}
