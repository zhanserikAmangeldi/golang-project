package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func Connect(url string) *sql.DB {
	db, err := sql.Open("postgres", url)
	if err != nil {
		log.Fatalf("failed open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("failed ping db: %v", err)
	}
	log.Println("connected to db")
	return db
}
