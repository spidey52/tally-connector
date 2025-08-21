package tally

import (
	"context"
	"fmt"
	"log"
	"strings"
	"tally-connector/cmd/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

type TallyPostgres struct {
	conn *pgx.Conn
}

func NewTallyPostgresDB(conn *pgx.Conn) *TallyPostgres {
	return &TallyPostgres{conn: conn}
}

func (db *TallyPostgres) CountDocuments(ctx context.Context, table_name string) (int, error) {
	var count int
	err := db.conn.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table_name)).Scan(&count)
	return count, err
}

func (db *TallyPostgres) LedgerVouchers(ctx context.Context, ledger_name string) ([]models.TrnVoucher, error) {

	var name string
	var alias string

	err := db.conn.QueryRow(ctx, "SELECT name, alias FROM mst_ledger WHERE name = $1 OR alias = $1", ledger_name).Scan(&name, &alias)
	if err != nil {
		log.Println("Error scanning ledger:", err)
		return nil, err
	}

	var vouchers []models.TrnVoucher
	var count int64

	var fields = strings.Join([]string{"guid", "alterid", "date", "party_name", "voucher_type", "voucher_number", "narration"}, ", ")
	var limit = "LIMIT 10"
	var filters = "WHERE party_name = $1"

	var query = fmt.Sprintf("SELECT %s FROM trn_voucher %s %s", fields, filters, limit)
	var countQuery = fmt.Sprintf("SELECT COUNT(*) FROM trn_voucher %s", filters)

	err = pgxscan.Select(ctx, db.conn, &vouchers, query, name)

	if err != nil {
		log.Println("Error querying vouchers:", err)
		return nil, err
	}

	err = db.conn.QueryRow(ctx, countQuery, name).Scan(&count)

	if err != nil {
		fmt.Println("Error querying vouchers count:", err)
		return nil, err
	}

	fmt.Println("Total vouchers found:", count)

	return vouchers, nil

}

// func (db *TallyPostgres) LedgerDetails(ctx context.Context, ledgerID string) (VoucherDetails, error) {
// 	var inventory []models.InventoryItem

// 	pgxscan.Select(ctx, db.conn, &inventory, `SELECT item, quantity, rate, amount, additional_amount, discount_amount
// 	FROM trn_inventory
// 	WHERE guid = $1
// 	`, ledgerID)

// 	var inv []models.TrnInventoryAccounting

// 	pgxscan.Select(ctx, db.conn, &inv, `SELECT guid, item, quantity, rate, amount, additional_amount, discount_amount
// 	FROM trn_inventory_accounting
// 	WHERE guid = $1
// 	`, ledgerID)

// 	var accounting []models.TrnAccounting

// 	pgxscan.Select(ctx, db.conn, &accounting, `SELECT guid, ledger, amount
// 	FROM trn_accounting
// 	WHERE guid = $1
// 	`, ledgerID)

// 	return VoucherDetails{
// 		Inventory:           inventory,
// 		InventoryAccounting: inv,
// 		Accounting:          accounting,
// 	}, nil

// }
