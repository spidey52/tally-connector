package models

import "time"

type SyncTable struct {
	Name         string `json:"name" db:"table_name"`
	Group        string `json:"group" db:"group_name"`
	MaxWait      int    `json:"max_wait" db:"max_wait"`           // in seconds
	SyncInterval int    `json:"sync_interval" db:"sync_interval"` // in minutes
}

type SyncLog struct {
	Table     string    `json:"table" db:"table_name"`
	Group     string    `json:"group" db:"group_name"`
	StartTime time.Time `json:"start_time" db:"start_time"`
	EndTime   time.Time `json:"end_time" db:"end_time"`
	Duration  int       `json:"duration" db:"duration"` // in seconds
	Status    string    `json:"status" db:"status"`
	Message   string    `json:"message" db:"message"`
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
    status varchar(50) not null
  );

*/
