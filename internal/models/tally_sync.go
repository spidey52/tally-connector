package models

import (
	"context"
	"log"
	"tally-connector/internal/db"
	"time"
)

type SyncTable struct {
	Name         string `json:"name" db:"table_name"`
	Group        string `json:"group" db:"group_name"`
	MaxWait      int    `json:"max_wait" db:"max_wait"`           // in seconds
	SyncInterval int    `json:"sync_interval" db:"sync_interval"` // in minutes

	LastSyncTime time.Time `json:"last_sync_time,omitempty" db:"-"`
}

type SyncLog struct {
	Table     string    `json:"table" db:"table_name"`
	Group     string    `json:"group" db:"group_name"`
	StartTime time.Time `json:"start_time" db:"start_time"`
	EndTime   time.Time `json:"end_time" db:"end_time"`
	Duration  float64   `json:"duration" db:"duration"` // in seconds
	Status    string    `json:"status" db:"status"`
	Message   string    `json:"message" db:"message"`
}

func CreateSyncTables() {

	val, err := db.GetDB().Exec(context.Background(), `
    create table if not exists tbl_sync_tables (
      table_name varchar(255) not null,
      group_name varchar(255) not null,
      max_wait integer not null,
      sync_interval integer not null
    )`)

	if err != nil {
		// Handle error
		log.Println("Error creating tbl_sync_tables:", err)
		return
	}

	log.Println("Result of creating tbl_sync_tables:", val)

	val, err = db.GetDB().Exec(context.Background(), `
    create table if not exists tbl_sync_logs (
      table_name varchar(255) not null,
      group_name varchar(255) not null,
      start_time timestamp not null,
      end_time timestamp not null,
      duration float not null,
      status varchar(50) not null,
      message text not null default ''
    );`)

	if err != nil {
		// Handle error
		log.Println("Error creating tbl_sync_logs:", err)
		return
	}

	log.Println("Result of creating tbl_sync_logs:", val)

}

/*

  create table if not exists tbl_sync_tables (
    name varchar(255) not null,
    group_name varchar(255) not null,
    max_wait integer not null,
    sync_interval integer not null
  );

  create table if not exists tbl_sync_logs (
    table_name varchar(255) not null,
    group_name varchar(255) not null,
    start_time timestamp not null,
    end_time timestamp not null,
    duration integer not null,
    status varchar(50) not null,
    message text not null default ''
  );

*/
