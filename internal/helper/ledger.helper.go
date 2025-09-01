package helper

import (
	"context"
	"tally-connector/internal/db"
	"tally-connector/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
)

// map of [alias || name] => name

func LedgerMapByAlias() map[string]string {
	ledgerMap := make(map[string]string)

	var ledgers = []models.Ledger{}

	err := pgxscan.Select(context.Background(), db.GetDB(), &ledgers, "select name, alias from mst_ledger ")

	if err != nil {
		return ledgerMap
	}

	for _, ledger := range ledgers {
		ledgerMap[ledger.Alias] = ledger.Name
		ledgerMap[ledger.Name] = ledger.Name
	}

	return ledgerMap

}
