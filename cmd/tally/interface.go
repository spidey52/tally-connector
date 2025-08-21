package tally

import (
	"tally-connector/cmd/models"
)

type VoucherDetails struct {
	InventoryAccounting []models.TrnInventoryAccounting
	Accounting          []models.TrnAccounting
}

// type TallyDB interface {
// 	CountDocuments(ctx context.Context, table_name string) (int, error)
// 	LedgerVouchers(ctx context.Context, ledgerID string) ([]models.TrnVoucher, error)
// 	LedgerDetails(ctx context.Context, ledgerID string) (VoucherDetails, error)
// }
