package config

import (
	"database/sql"
	"log"

	"github.com/deside01/tg_freelance_bot/internal/database"
	_ "modernc.org/sqlite"
)

var DB *database.Queries

func SetupDB() {
	db, err := sql.Open("sqlite", "sqlite:../../mydb.db")
		if err != nil {
		log.Fatal(err)
	}

	DB = database.New(db)
}