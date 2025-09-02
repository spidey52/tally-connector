package db

import (
	"context"
	"log"
	"slices"
	"tally-connector/config"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
)

func SeedSyncTables(ctx context.Context) {

	tables, err := config.GetMergedTables()

	if err != nil {
		return
	}

	table_names := make([]string, len(tables))
	for i, table := range tables {
		table_names[i] = table.Name
	}

	// existing tables.
	var existing_tables []string
	err = pgxscan.Select(ctx, GetDB(), &existing_tables, "SELECT table_name FROM tbl_sync_tables")

	if err != nil {
		log.Printf("Error fetching existing tables: %v", err)
		return
	}

	to_insert := []string{}
	for _, table := range tables {
		if !slices.Contains(existing_tables, table.Name) {
			to_insert = append(to_insert, table.Name)
		}
	}

	if len(to_insert) == 0 {
		log.Println("No new tables to insert.")
		return
	}

	q := psql.Insert(im.Into("tbl_sync_tables"))

	for _, table := range to_insert {
		q.Apply(
			im.Values(psql.Arg(table, "group", 300, 5)),
		)
	}

	query, args, err := q.Build(ctx)

	if err != nil {
		log.Printf("Error building query: %v", err)
		return
	}

	log.Println("Built query:", query)

	_, err = GetDB().Exec(ctx, query, args...)

	if err != nil {
		log.Printf("Error seeding sync tables: %v", err)
		return
	}

	log.Printf("Seeded %d new tables into tbl_sync_tables", len(to_insert))

}
