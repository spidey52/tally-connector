package main

import (
	"tally-connector/cmd/loader/config"
)

func SeedSyncTables() {

	tables, err := config.GetMergedTables()

	if err != nil {
		return
	}

	table_names := make([]string, len(tables))
	for i, table := range tables {
		table_names[i] = table.Name
	}

}
