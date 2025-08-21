package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

var db *pgx.Conn

var connStr = "postgres://satyam:satyam52@localhost:5432/tally_db"

func ConnectDB() {
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal("Unable to connect: ", err)
	}

	db = conn
}

func GetDB() *pgx.Conn {
	if db == nil {
		ConnectDB()
	}
	return db
}
