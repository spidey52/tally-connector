package main

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"tally-connector/cmd/db"
	"tally-connector/cmd/loader/config"
	"time"

	"github.com/clbanning/mxj/v2"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// The InsertIntoDB method updated to use pgx's CopyFrom for bulk insertion.
// It assumes a db connection pool is available.
func InsertIntoDB(ctx context.Context, table_name string, cols []string, data []map[string]any) error {
	tx, err := db.GetDB().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Step 1: Truncate the table to delete all rows.
	// We use TRUNCATE instead of DELETE for better performance on large tables.
	// The Sanitize() method protects against SQL injection for the table name.
	truncateStmt := fmt.Sprintf("TRUNCATE TABLE %s", pgx.Identifier{table_name}.Sanitize())
	_, err = tx.Exec(ctx, truncateStmt)
	if err != nil {
		return fmt.Errorf("failed to truncate table %s: %w", table_name, err)
	}

	// Step 2: Use CopyFrom for a fast, bulk insert.
	source := &mapSource{
		data: data,
		cols: cols,
		idx:  -1,
	}

	tableName := pgx.Identifier{table_name}
	_, err = tx.CopyFrom(ctx, tableName, cols, source)
	if err != nil {
		return fmt.Errorf("CopyFrom failed: %w", err)
	}

	// Step 3: Commit the entire transaction.
	// This makes both the TRUNCATE and the bulk insert permanent.
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	fmt.Printf("Successfully truncated all rows and inserted %d new rows into %s.\n", len(data), table_name)
	return nil
}

// mapSource is a helper struct that implements the pgx.CopyFromSource interface.
type mapSource struct {
	data []map[string]any
	cols []string
	idx  int
}

func (s *mapSource) Next() bool {
	s.idx++
	return s.idx < len(s.data)
}

func (s *mapSource) Values() ([]any, error) {
	row := s.data[s.idx]
	values := make([]any, 0, len(s.cols))
	for _, col := range s.cols {
		values = append(values, row[col])
	}
	return values, nil
}

func (s *mapSource) Err() error {
	return nil
}

// ProcessXMLData extracts and transforms XML data into a JSON array of objects.
func ProcessXMLData(c *gin.Context, table string, result string, fields []config.Field) {
	m, err := mxj.NewMapXml([]byte(result))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse XML", "details": err.Error()})
		return
	}

	realData, ok := m["ENVELOPE"].(map[string]any)
	if !ok {
		c.JSON(500, gin.H{"error": "XML structure missing 'ENVELOPE' key"})
		return
	}

	// This part is the core of the transformation.
	// We'll use a direct approach rather than fragile key sorting.
	keys := make([]string, 0, len(realData))
	for key := range realData {
		keys = append(keys, key)
	}
	sort.Strings(keys) // Sorting keys for consistent order, assuming data arrays have same length

	// Check if there is at least one key to iterate over
	if len(keys) == 0 {
		c.JSON(200, gin.H{"data": []map[string]any{}})
		return
	}

	// Use the length of the first sorted key's data array to control the loop
	firstKeyData, ok := realData[keys[0]].([]any)
	if !ok {
		c.JSON(500, gin.H{"error": "Data for first key is not a slice"})
		return
	}
	dataLength := len(firstKeyData)

	finalResult := make([]map[string]any, 0, dataLength)

	cols := make([]string, 0, len(fields))
	for _, field := range fields {
		cols = append(cols, field.Name)
	}

	// delete blank key from realData and checkif len(realData = 0) return 500

	delete(realData, "BLANK")

	if len(realData) == 0 {
		c.JSON(500, gin.H{"error": "no valid data found"})
		return
	}

	if len(realData) != len(cols) {
		c.JSON(500, gin.H{"error": "data length mismatch",
			"expected": len(cols),
			"actual":   len(realData),
		})
		return
	}

	// Iterate based on the number of records
	for i := range dataLength {
		record := make(map[string]any)
		// Populate the record by iterating through all keys
		for idx, key := range keys {
			field := fields[idx]
			assignKey := field.Name

			keyData, ok := realData[key].([]any)
			if !ok || i >= len(keyData) {
				// Handle cases where a key's data is not a slice or is too short
				record[key] = nil // Or some default value
				continue
			}
			val, err := ConvertDataType(keyData[i].(string), field.Type)
			if err != nil {
				record[assignKey] = nil
			} else {
				record[assignKey] = val
			}

		}
		finalResult = append(finalResult, record)
	}

	ctx := c.Request.Context()

	err = InsertIntoDB(ctx, table, cols, finalResult)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to insert data into database", "details": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"data":       finalResult[0],
		"count":      len(finalResult),
		"cols":       cols,
		"datalength": len(firstKeyData),
	})
}

func ConvertDataType(value string, targetType string) (any, error) {
	if value == "ñ" {
		return nil, nil
	}

	if targetType == "text" {
		return value, nil
	}

	if targetType == "number" || targetType == "logical" || targetType == "amount" || targetType == "quantity" || targetType == "rate" {
		return ConvertToNumber(value)
	}

	if targetType == "date" {
		return ConvertToDate(value)
	}

	return value, fmt.Errorf("unsupported conversion from %T to %s", value, targetType)
}

func ConvertToNumber(value string) (float64, error) {

	if num, err := strconv.ParseFloat(value, 64); err == nil {
		return num, nil
	}
	return 0, fmt.Errorf("invalid number: %s", value)
}

func ConvertToDate(value string) (time.Time, error) {

	if date, err := time.Parse("2006-01-02", value); err == nil {
		return date, nil
	}
	return time.Time{}, fmt.Errorf("invalid date: %s", value)
}

func ConvertLogical(value string) int {
	if value == "1" {
		return 1
	} else {
		return 0
	}
}
