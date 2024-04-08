package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

const (
	CreateScript   = "create"
	DeleteScript   = "delete"
	TruncateScript = "truncate"
)

func ExecuteScript(db *sql.DB, scriptName string) {
	sqlFile, err := os.ReadFile(fmt.Sprintf("./db/sql/%s.sql", scriptName))
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(string(sqlFile))
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Successful executed script: %s", scriptName)
	}
}

func Init(db *sql.DB) {
	ExecuteScript(db, CreateScript)
}

func ShutDown(db *sql.DB) {
	ExecuteScript(db, TruncateScript) // Fast Path
	ExecuteScript(db, DeleteScript)
}

func AddPlayer(db *sql.DB, username string) {
	db.Query("INSERT INTO players(username, score) VALUES($1, 0)", username)
}

func CheckPlayerExistence(db *sql.DB, username string) bool {
	rows, err := db.Query("SELECT * FROM players WHERE username = $1", username)
	if err != nil {
		log.Println(err)
	}
	return rows.Next()
}

func AddWord(db *sql.DB, word string, from string) {
	db.Query("INSERT INTO session_words(word, username) VALUES($1, $2)", word, from)
	db.Query("UPDATE players SET score = score + 1 WHERE username = $1", from) // Postgres trigger???
}

func GetLastWord(db *sql.DB) string {
	var result string
	err := db.QueryRow("SELECT word FROM session_words WHERE id = (SELECT MAX(id) FROM session_words);").Scan(&result)
	if err != nil {
		log.Println(err)
	}
	return result
}
