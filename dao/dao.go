package dao

import (
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

func AddPlayer(db *gorm.DB, id int64, username string, firstName string) {
	db.Create(&Player{ID: id, Username: username, Score: 0, FirstName: firstName})
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

func AddWord(db *gorm.DB, word string, kana string, username string, userID int64) {
	db.Create(&Word{UserID: userID, Username: username, Word: word, Kana: kana})
	db.Model(&Player{}).Where("username = ?", username).
		Update("score", gorm.Expr("score + ?", 1))
}

func SetScore(db *gorm.DB, player string, to uint64) {
	db.Model(&Player{}).Where("username = ?", player).
		Update("score", to)
}

func LastWord(db *gorm.DB) (string, string) {
	var lastWord Word
	db.Last(&lastWord)
	return lastWord.Word, lastWord.Kana
}

func LastPlayer(db *gorm.DB) int64 {
	var lastWord Word
	db.Last(&lastWord)
	return lastWord.UserID
}

func CheckWordExistence(db *gorm.DB, word string) bool {
	var words []Word
	db.Model(&Word{}).Where("word = ?", word).Find(&words)
	return len(words) != 0
}
