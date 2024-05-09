package dao

import (
	"gorm.io/gorm"
)

type Word struct {
	ID       int `gorm:"primaryKey;autoIncrement"`
	Word     string
	Kana     string
	Username string
}

type Player struct {
	ID        int    `gorm:"primaryKey;autoIncrement"`
	FirstName string //Pretty stats at the end
	Username  string
	Score     uint64
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

func AddPlayer(db *gorm.DB, username string, firstName string) {
	db.Create(&Player{Username: username, Score: 0, FirstName: firstName})
}

func AllPlayers(db *gorm.DB) []Player {
	var players []Player
	db.Model(&Player{}).Find(&players)
	return players
}

func CheckPlayerExistence(db *gorm.DB, username string) bool {
	var players []Player
	db.Model(&Player{}).Where("username = ?", username).Find(&players)
	return len(players) != 0
}

func AddWord(db *gorm.DB, word string, kana string, from string) {
	db.Create(&Word{Username: from, Word: word, Kana: kana})
	db.Model(&Player{}).Where("username = ?", from).
		Update("score", gorm.Expr("score + ?", 1))
}

func LastWord(db *gorm.DB) (string, string) {
	var lastWord Word
	db.Last(&lastWord)
	return lastWord.Word, lastWord.Kana
}

func CheckWordExistence(db *gorm.DB, word string) bool {
	var words []Word
	db.Model(&Word{}).Where("word = ?", word).Find(&words)
	return len(words) != 0
}
